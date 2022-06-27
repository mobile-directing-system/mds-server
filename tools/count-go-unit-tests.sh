BASE_DIR=".."

GO_SERVICES="$BASE_DIR/services/go"

TAB="	"

MATCH_ROOT_TEST="func Test.+\(.+ \*testing\.T\)"
MATCH_SUITES="${TAB}suite.Suite"
MATCH_SUITE_TESTS="func \(.+\) Test.*\(\) {"

INCLUDE="*_test.go"

root_tests=$(grep -r --include="${INCLUDE}" -E "${MATCH_ROOT_TEST}" $GO_SERVICES | wc -l)
suites=$(grep -r --include="${INCLUDE}" -E "${MATCH_SUITES}" $GO_SERVICES | wc -l)
suite_tests=$(grep -r --include="${INCLUDE}" -E "${MATCH_SUITE_TESTS}" $GO_SERVICES | wc -l)

test_count=$((root_tests - suites + suite_tests))

echo "$test_count"