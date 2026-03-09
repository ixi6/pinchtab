#!/bin/bash
# 04-tabs-api.sh — Tab-scoped API tests (regression test for #207)

source "$(dirname "$0")/common.sh"

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab snap --tab <id> (regression #207)"

# Create a tab
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/index.html\"}"
CREATED_TAB=$(echo "$RESULT" | jq -r '.tabId')
echo -e "  Created tab: ${CREATED_TAB:0:12}..."

# Get tab ID from /tabs endpoint
pt_get /tabs
LISTED_TAB=$(echo "$RESULT" | jq -r '.tabs[-1].id')
echo -e "  Listed tab:  ${LISTED_TAB:0:12}..."

# Test: /tabs/{id}/snapshot should work (was broken in #207)
pt_get "/tabs/${LISTED_TAB}/snapshot"
assert_status 200 "${PINCHTAB_URL}/tabs/${LISTED_TAB}/snapshot"
assert_json_contains "$RESULT" '.title' 'E2E Test'

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab text/screenshot --tab <id>"

pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/form.html\"}"
TAB_ID=$(echo "$RESULT" | jq -r '.tabId')

# Test /tabs/{id}/text
assert_status 200 "${PINCHTAB_URL}/tabs/${TAB_ID}/text"

# Test /tabs/{id}/screenshot
assert_status 200 "${PINCHTAB_URL}/tabs/${TAB_ID}/screenshot"

end_test

# ─────────────────────────────────────────────────────────────────
start_test "pinchtab tab close"

BEFORE=$(get_tab_count)

# Create and close a tab
pt_post /navigate -d "{\"url\":\"${FIXTURES_URL}/buttons.html\"}"
TAB_ID=$(echo "$RESULT" | jq -r '.tabId')

assert_status 200 "${PINCHTAB_URL}/tabs/${TAB_ID}/close" "POST"

sleep 1
assert_tab_closed "$BEFORE"

end_test
