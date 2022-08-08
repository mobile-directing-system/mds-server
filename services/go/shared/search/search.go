// Package search bundles common search functionality for Meilisearch.
package search

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/lefinal/meh"
	"github.com/meilisearch/meilisearch-go"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
	"reflect"
	"time"
)

const awaitTaskCooldown = 400 * time.Millisecond

// DefaultClientTimeout to use for Client when creating via ClientConfig.
const DefaultClientTimeout = 10 * time.Second

// Index is the UID of a Meilisearch index.
type Index string

// Attribute is a typed attribute name.
type Attribute string

// AttributeSet is a typed attribute set.
type AttributeSet map[Attribute]any

// Document is a typed string-any-map.
type Document map[Attribute]any

// IndexConfig is a typed meilisearch.IndexConfig.
type IndexConfig struct {
	// PrimaryKey is the typed version in meilisearch.IndexConfig.
	PrimaryKey Attribute
	// Searchable is the typed version in meilisearch.IndexConfig.
	Searchable []Attribute
	// Ranking is the typed version in meilisearch.IndexConfig.
	Ranking []string
	// Filterable is the typed version in meilisearch.IndexConfig.
	Filterable []Attribute
	// Sortable is the typed version in meilisearch.IndexConfig.
	Sortable []Attribute
}

// ClientConfig is used for Launch of a new Client.
type ClientConfig struct {
	// Host under which Meilisearch is accessible.
	Host         string
	MasterKey    string
	IndexConfigs map[Index]IndexConfig
	Logger       *zap.Logger
	Timeout      time.Duration
}

// Client allows searching and setting up/manipulating indices.
type Client interface {
	Index(index Index) meilisearch.IndexInterface
	CreateIndex(config *meilisearch.IndexConfig) (resp *meilisearch.TaskInfo, err error)
	// IndexConfig returns the IndexConfig for the given Index. If the Index is
	// unknown, a meh.ErrInternal will be returned.
	IndexConfig(index Index) (IndexConfig, error)
	GetIndex(uid string) (resp *meilisearch.Index, err error)
	Logger() *zap.Logger
}

// client is the actual implemenation of Client.
type client struct {
	// msClient is the meilisearch.Client to use and should never be nil.
	msClient     *meilisearch.Client
	logger       *zap.Logger
	indexConfigs map[Index]IndexConfig
}

func (c *client) Index(index Index) meilisearch.IndexInterface {
	return c.msClient.Index(string(index))
}

func (c *client) Logger() *zap.Logger {
	if c.logger == nil {
		return zap.NewNop()
	}
	return c.logger
}

func (c *client) IndexConfig(index Index) (IndexConfig, error) {
	config, ok := c.indexConfigs[index]
	if !ok {
		knownIndexes := make([]Index, 0, len(c.indexConfigs))
		for i := range c.indexConfigs {
			knownIndexes = append(knownIndexes, i)
		}
		return IndexConfig{}, meh.NewInternalErr("index not found", meh.Details{"known": knownIndexes})
	}
	return config, nil
}

func (c *client) CreateIndex(config *meilisearch.IndexConfig) (resp *meilisearch.TaskInfo, err error) {
	return c.msClient.CreateIndex(config)
}

func (c *client) GetIndex(uid string) (resp *meilisearch.Index, err error) {
	return c.msClient.GetIndex(uid)
}

// Launch creates and prepares a nwe Client with the given ClientConfig. This
// may be a long-running operation because we wait for index updates.
func Launch(ctx context.Context, clientConfig ClientConfig) (Client, error) {
	if clientConfig.Timeout == 0 {
		clientConfig.Timeout = DefaultClientTimeout
	}
	msClient := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    clientConfig.Host,
		APIKey:  clientConfig.MasterKey,
		Timeout: clientConfig.Timeout,
	})
	c := &client{
		msClient:     msClient,
		logger:       clientConfig.Logger,
		indexConfigs: clientConfig.IndexConfigs,
	}
	err := launchClient(ctx, c, clientConfig)
	if err != nil {
		return nil, meh.Wrap(err, "launch client", nil)
	}
	return c, nil
}

func launchIndex(ctx context.Context, c Client, index Index, indexConfig IndexConfig) error {
	// Assure index exists.
	_, err := c.GetIndex(string(index))
	if err != nil {
		switch e := err.(type) {
		case *meilisearch.Error:
			if e.StatusCode == http.StatusNotFound {
				// Create index.
				task, err := c.CreateIndex(&meilisearch.IndexConfig{
					Uid:        string(index),
					PrimaryKey: string(indexConfig.PrimaryKey),
				})
				if err == nil {
					c.Logger().Debug("awaiting index creation",
						zap.Any("index", index),
						zap.Any("task_uid", task.TaskUID))
					err = awaitTask(ctx, c, index, task.TaskUID)
					if err != nil {
						return meh.Wrap(err, "await create-index-task", meh.Details{"task_uid": task.TaskUID})
					}
					c.Logger().Debug("index created",
						zap.Any("index", index),
						zap.Any("task_uid", task.TaskUID))
				}
			}
		default:
			return meh.NewInternalErrFromErr(err, "get index", nil)
		}
	}
	// Update settings.
	err = updateIndexSettings(ctx, c, index, indexConfig)
	if err != nil {
		return meh.Wrap(err, "update index settings", meh.Details{"index": index})
	}
	return nil
}

// launchClient prepares the given Client with the ClientConfig.
func launchClient(ctx context.Context, c Client, config ClientConfig) error {
	// Update index settings.
	eg, egCtx := errgroup.WithContext(ctx)
	for iindex := range config.IndexConfigs {
		index := iindex
		indexConfig := config.IndexConfigs[iindex]
		eg.Go(func() error {
			err := launchIndex(egCtx, c, index, indexConfig)
			if err != nil {
				return meh.Wrap(err, "launch index", meh.Details{
					"index":        index,
					"index_config": indexConfig,
				})
			}
			return nil
		})
	}
	err := eg.Wait()
	if err != nil {
		return err
	}
	return nil
}

// updateIndexSettings updates changed settings for the give Index and waits for
// task completion.
func updateIndexSettings(ctx context.Context, c Client, index Index, config IndexConfig) error {
	// Retrieve current settings.
	currentSettings, err := c.Index(index).GetSettings()
	if err != nil {
		return meh.NewInternalErrFromErr(err, "get settings for index", nil)
	}
	// Calculate new settings.
	newSettings := genMSSettings(config)
	cleanedSettings := removeUnchangedFromMSSettings(*currentSettings, newSettings)
	// Update settings.
	c.Logger().Debug("updating index settings",
		zap.Any("config_in", config),
		zap.Any("current_settings", *currentSettings),
		zap.Any("settings_from_config", newSettings),
		zap.Any("settings_cleaned", cleanedSettings))
	task, err := c.Index(index).UpdateSettings(&cleanedSettings)
	if err != nil {
		return meh.NewInternalErrFromErr(err, "update settings", meh.Details{"new_settings": cleanedSettings})
	}
	err = awaitTask(ctx, c, index, task.TaskUID)
	if err != nil {
		return meh.Wrap(err, "await update-settings-task", meh.Details{
			"index":    index,
			"task_uid": task.TaskUID,
		})
	}
	return nil
}

// awaitTask awaits the task with the given UID to finish. For cooldown, it uses
// awaitTaskCooldown. If the task fails, the error is returned as well including
// error details by Meilisearch.
func awaitTask(ctx context.Context, c Client, index Index, taskUID int64) error {
	for {
		taskInfo, err := c.Index(index).GetTask(taskUID)
		if err != nil {
			return meh.NewInternalErrFromErr(err, "get task", meh.Details{"task_uid": taskUID})
		}
		if taskInfo.Status == meilisearch.TaskStatusSucceeded {
			return nil
		}
		if taskInfo.Status == meilisearch.TaskStatusFailed {
			e := taskInfo.Error
			return meh.Wrap(meh.NewInternalErr(e.Message, meh.Details{
				"code": e.Code,
				"type": e.Type,
				"link": e.Link,
			}), "task failed", nil)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(awaitTaskCooldown):
		}
	}
}

func genMSSettings(config IndexConfig) meilisearch.Settings {
	distinctAttribute := &config.PrimaryKey
	searchableAttributes := make([]string, 0, len(config.Searchable))
	for _, attribute := range config.Searchable {
		searchableAttributes = append(searchableAttributes, string(attribute))
	}
	filterableAttributes := make([]string, 0, len(config.Filterable))
	for _, attribute := range config.Filterable {
		filterableAttributes = append(filterableAttributes, string(attribute))
	}
	sortableAttributes := make([]string, 0, len(config.Sortable))
	for _, attribute := range config.Sortable {
		sortableAttributes = append(sortableAttributes, string(attribute))
	}
	return meilisearch.Settings{
		RankingRules:         config.Ranking,
		DistinctAttribute:    (*string)(distinctAttribute),
		SearchableAttributes: searchableAttributes,
		DisplayedAttributes:  []string{string(config.PrimaryKey)},
		StopWords:            nil,
		Synonyms:             nil,
		FilterableAttributes: filterableAttributes,
		SortableAttributes:   sortableAttributes,
		TypoTolerance:        nil,
		Pagination:           nil,
		Faceting:             nil,
	}
}

// removeUnchangedFromMSSettings removed most of the unchanged settings in
// meilisearch.Settings in order to reduce update-time.
func removeUnchangedFromMSSettings(current meilisearch.Settings, new meilisearch.Settings) meilisearch.Settings {
	// Helper fn.
	shouldUpdateStringSlice := func(current []string, new []string) bool {
		if new == nil {
			return false
		}
		if len(current) != len(new) {
			return true
		}
		for i := range current {
			if current[i] != new[i] {
				return true
			}
		}
		return true
	}
	shouldUpdateStringPtr := func(current *string, new *string) bool {
		if new == nil {
			return false
		}
		if current == nil {
			return true
		}
		// Compare contents.
		if *current != *new {
			return true
		}
		return true
	}
	shouldUpdateStringSliceMap := func(current map[string][]string, new map[string][]string) bool {
		if new == nil {
			return false
		}
		// Copy current.
		currentCopy := make(map[string][]string, len(current))
		for key, val := range current {
			vCopy := make([]string, 0, len(val))
			for _, s := range val {
				vCopy = append(vCopy, s)
			}
			currentCopy[key] = vCopy
		}
		// Compare.
		for keyNew, valNew := range new {
			vCurrent, ok := currentCopy[keyNew]
			if !ok || valNew == nil {
				return true
			}
			if shouldUpdateStringSlice(vCurrent, valNew) {
				return true
			}
			delete(currentCopy, keyNew)
		}
		return false
	}
	// Clean.
	cleaned := meilisearch.Settings{}
	if shouldUpdateStringSlice(current.RankingRules, new.RankingRules) {
		cleaned.RankingRules = new.RankingRules
	}
	if shouldUpdateStringPtr(current.DistinctAttribute, new.DistinctAttribute) {
		cleaned.DistinctAttribute = new.DistinctAttribute
	}
	if shouldUpdateStringSlice(current.SearchableAttributes, new.SearchableAttributes) {
		cleaned.SearchableAttributes = new.SearchableAttributes
	}
	if shouldUpdateStringSlice(current.DisplayedAttributes, new.DisplayedAttributes) {
		cleaned.DisplayedAttributes = new.DisplayedAttributes
	}
	if shouldUpdateStringSlice(current.StopWords, new.StopWords) {
		cleaned.StopWords = new.StopWords
	}
	if shouldUpdateStringSliceMap(current.Synonyms, new.Synonyms) {
		cleaned.Synonyms = new.Synonyms
	}
	if shouldUpdateStringSlice(current.FilterableAttributes, new.FilterableAttributes) {
		cleaned.FilterableAttributes = new.FilterableAttributes
	}
	if new.TypoTolerance != nil {
		cleaned.TypoTolerance = new.TypoTolerance
	}
	if new.Pagination != nil {
		cleaned.Pagination = new.Pagination
	}
	if new.Faceting != nil {
		cleaned.Faceting = new.Faceting
	}
	return cleaned
}

// Result represents meilisearch.SearchResponse but with typed entries.
type Result[T any] struct {
	Hits               []T           `json:"hits"`
	EstimatedTotalHits int           `json:"estimated_total_hits"`
	Offset             int           `json:"offset"`
	Limit              int           `json:"limit"`
	ProcessingTime     time.Duration `json:"processing_time"`
	Query              string        `json:"query"`
}

// ResultFromResult copies the given Result but sets Result.Hits to the new
// given ones.
func ResultFromResult[From any, To any](from Result[From], newHits []To) Result[To] {
	return Result[To]{
		Hits:               newHits,
		EstimatedTotalHits: from.EstimatedTotalHits,
		Offset:             from.Offset,
		Limit:              from.Limit,
		ProcessingTime:     from.ProcessingTime,
		Query:              from.Query,
	}
}

// MapResult maps types for the given Result.
func MapResult[From any, To any](from Result[From], mapFn func(From) To) Result[To] {
	mappedHits := make([]To, 0, len(from.Hits))
	for _, f := range from.Hits {
		mappedHits = append(mappedHits, mapFn(f))
	}
	return Result[To]{
		Hits:               mappedHits,
		EstimatedTotalHits: from.EstimatedTotalHits,
		Offset:             from.Offset,
		Limit:              from.Limit,
		ProcessingTime:     from.ProcessingTime,
		Query:              from.Query,
	}
}

// UUIDSearch searches the given Index with Params and parses the returned ID
// based on the IndexConfig.PrimaryKey for the given Index.
func UUIDSearch(c Client, index Index, searchParams Params) (Result[uuid.UUID], error) {
	msResult, err := c.Index(index).Search(searchParams.Query, &meilisearch.SearchRequest{
		Offset: int64(searchParams.Offset),
		Limit:  int64(searchParams.Limit),
	})
	if err != nil {
		return Result[uuid.UUID]{}, meh.NewInternalErrFromErr(err, "search", meh.Details{
			"query":  searchParams.Query,
			"offset": searchParams.Offset,
			"limit":  searchParams.Limit,
		})
	}
	// Parse all UUIDs.
	result := Result[uuid.UUID]{
		Hits:               make([]uuid.UUID, 0, len(msResult.Hits)),
		EstimatedTotalHits: int(msResult.EstimatedTotalHits),
		Offset:             int(msResult.Offset),
		Limit:              int(msResult.Limit),
		ProcessingTime:     time.Duration(msResult.ProcessingTimeMs) * time.Millisecond,
		Query:              msResult.Query,
	}
	indexConfig, err := c.IndexConfig(index)
	if err != nil {
		return Result[uuid.UUID]{}, meh.Wrap(err, "get index config", nil)
	}
	for _, hit := range msResult.Hits {
		e, ok := hit.(map[string]any)
		if !ok {
			return Result[uuid.UUID]{}, meh.NewInternalErr("cannot cast hit to string",
				meh.Details{"was": reflect.TypeOf(hit)})
		}
		idRaw, ok := e[string(indexConfig.PrimaryKey)]
		if !ok {
			return Result[uuid.UUID]{}, meh.NewInternalErr("primary index field not found", meh.Details{
				"key": indexConfig.PrimaryKey,
				"hit": e,
			})
		}
		idStr, ok := idRaw.(string)
		if !ok {
			return Result[uuid.UUID]{}, meh.NewInternalErr("cannot cast field to string", meh.Details{
				"key":   indexConfig.PrimaryKey,
				"hit":   e,
				"idRaw": idRaw,
				"was":   reflect.TypeOf(idRaw),
			})
		}
		id, err := uuid.FromString(idStr)
		if err != nil {
			return Result[uuid.UUID]{}, meh.NewInternalErr("parse uuid", meh.Details{
				"hit": e,
				"was": idStr,
			})
		}
		result.Hits = append(result.Hits, id)
	}
	return result, nil
}
