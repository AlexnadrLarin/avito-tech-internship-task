API_URL ?= http://localhost:8082

.PHONY: init prepare ab

init:
	mkdir -p ab-tests
	echo '{"team_name":"backend","members":[{"user_id":"u1","username":"Alice","is_active":true},{"user_id":"u2","username":"Bob","is_active":true},{"user_id":"u3","username":"Charlie","is_active":true}]}' > ab-tests/team_add.json
	echo '{"pull_request_id":"pr-constant","pull_request_name":"LoadTest","author_id":"u1"}' > ab-tests/pr_create.json
	echo '{"user_id":"u2","is_active":false}' > ab-tests/user_set_active.json
	echo '{"pull_request_id":"pr-constant"}' > ab-tests/pr_merge.json
	echo '{"pull_request_id":"pr-constant","old_user_id":"u2"}' > ab-tests/pr_reassign.json

prepare:
	for i in $$(seq 1 10); do \
		curl -s $(API_URL)/team/get?team_name=backend >/dev/null 2>&1 && break || sleep 1; \
	done
	curl -s -X POST $(API_URL)/team/add -H "Content-Type: application/json" -d @ab-tests/team_add.json >/dev/null
	curl -s -X POST $(API_URL)/pullRequest/create -H "Content-Type: application/json" -d @ab-tests/pr_create.json >/dev/null || true

ab:
	ab -n 18000 -c 1200 -p ab-tests/team_add.json -T application/json $(API_URL)/team/add
	ab -n 33000 -c 2000 "$(API_URL)/team/get?team_name=backend"
	ab -n 18000 -c 1200 -p ab-tests/user_set_active.json -T application/json $(API_URL)/users/setIsActive
	ab -n 18000 -c 1200 -p ab-tests/pr_create.json -T application/json $(API_URL)/pullRequest/create
	ab -n 18000 -c 1200 -p ab-tests/pr_merge.json -T application/json $(API_URL)/pullRequest/merge
	ab -n 18000 -c 1200 -p ab-tests/pr_reassign.json -T application/json $(API_URL)/pullRequest/reassign
	ab -n 42000 -c 2600 "$(API_URL)/users/getReview?user_id=u2"
	ab -n 48000 -c 3000 "$(API_URL)/users/getReviewsStats"
