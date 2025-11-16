API_URL ?= http://localhost:8080

.PHONY: init prepare ab

init:
	mkdir -p ab-tests
	echo '{"team_name":"backend","members":[{"user_id":"u1","username":"Alice","is_active":true},{"user_id":"u2","username":"Bob","is_active":true},{"user_id":"u3","username":"Charlie","is_active":true},{"user_id":"u4","username":"Dave","is_active":true}]}' > ab-tests/team_add.json
	@echo "Initial data files created"

prepare:
	@for i in $$(seq 1 10); do \
		curl -s $(API_URL)/api/v1/team/get?team_name=test >/dev/null 2>&1 && break || sleep 2; \
	done
	curl -X POST $(API_URL)/api/v1/team/add -H "Content-Type: application/json" -d @ab-tests/team_add.json || true
	@echo "Creating PRs for merge/reassign tests..."
	@for i in $$(seq 1 100); do \
		UNIQUE_ID=$$(date +%s)-$$$$-$$(shuf -i 10000-99999 -n 1)-$$i; \
		echo "{\"pull_request_id\":\"pr-prep-$$UNIQUE_ID\",\"pull_request_name\":\"Add search\",\"author_id\":\"u1\"}" | curl -s -X POST $(API_URL)/api/v1/pullRequest/create -H "Content-Type: application/json" -d @- >/dev/null || true; \
	done

ab:
	@echo "Running load tests with unique IDs..."
	@echo "Test 1: Create teams (500 requests, 20 concurrent)..."
	@seq 1 500 | xargs -P 20 -I {} sh -c 'UNIQUE_ID=$$(date +%s)-$$$$-$$(shuf -i 10000-99999 -n 1)-{}; echo "{\"team_name\":\"backend-$$UNIQUE_ID\",\"members\":[{\"user_id\":\"u-$$UNIQUE_ID-1\",\"username\":\"Alice-$$UNIQUE_ID\",\"is_active\":true},{\"user_id\":\"u-$$UNIQUE_ID-2\",\"username\":\"Bob-$$UNIQUE_ID\",\"is_active\":true},{\"user_id\":\"u-$$UNIQUE_ID-3\",\"username\":\"Charlie-$$UNIQUE_ID\",\"is_active\":true},{\"user_id\":\"u-$$UNIQUE_ID-4\",\"username\":\"Dave-$$UNIQUE_ID\",\"is_active\":true}]}" | curl -s -X POST $(API_URL)/api/v1/team/add -H "Content-Type: application/json" -d @- >/dev/null'
	@echo "Test 2: Get teams (3000 requests, 100 concurrent)..."
	ab -n 3000 -c 100 "$(API_URL)/api/v1/team/get?team_name=backend"
	@echo "Test 3: Set user active (2000 requests, 50 concurrent)..."
	@seq 1 2000 | xargs -P 50 -I {} sh -c 'echo "{\"user_id\":\"u3\",\"is_active\":false}" | curl -s -X POST $(API_URL)/api/v1/users/setIsActive -H "Content-Type: application/json" -d @- >/dev/null'
	@echo "Test 4: Create pull requests (1500 requests, 50 concurrent)..."
	@seq 1 1500 | xargs -P 50 -I {} sh -c 'UNIQUE_ID=$$(date +%s)-$$$$-$$(shuf -i 10000-99999 -n 1)-{}; echo "{\"pull_request_id\":\"pr-$$UNIQUE_ID\",\"pull_request_name\":\"Add search\",\"author_id\":\"u1\"}" | curl -s -X POST $(API_URL)/api/v1/pullRequest/create -H "Content-Type: application/json" -d @- >/dev/null'
	@echo "Test 5: Merge pull requests (5000 requests, 200 concurrent) - creating PRs first..."
	@seq 1 5000 | xargs -P 200 -I {} sh -c 'UNIQUE_ID=$$(date +%s)-$$$$-$$(shuf -i 10000-99999 -n 1)-{}; PR_ID="pr-$$UNIQUE_ID"; echo "{\"pull_request_id\":\"$$PR_ID\",\"pull_request_name\":\"Add search\",\"author_id\":\"u1\"}" | curl -s -X POST $(API_URL)/api/v1/pullRequest/create -H "Content-Type: application/json" -d @- >/dev/null; echo "{\"pull_request_id\":\"$$PR_ID\"}" | curl -s -X POST $(API_URL)/api/v1/pullRequest/merge -H "Content-Type: application/json" -d @- >/dev/null'
	@echo "Test 6: Reassign reviewers (3000 requests, 100 concurrent) - creating PRs first..."
	@seq 1 3000 | xargs -P 100 -I {} sh -c 'UNIQUE_ID=$$(date +%s)-$$$$-$$(shuf -i 10000-99999 -n 1)-{}; PR_ID="pr-$$UNIQUE_ID"; echo "{\"pull_request_id\":\"$$PR_ID\",\"pull_request_name\":\"Add search\",\"author_id\":\"u1\"}" | curl -s -X POST $(API_URL)/api/v1/pullRequest/create -H "Content-Type: application/json" -d @- >/dev/null; echo "{\"pull_request_id\":\"$$PR_ID\",\"old_user_id\":\"u2\"}" | curl -s -X POST $(API_URL)/api/v1/pullRequest/reassign -H "Content-Type: application/json" -d @- >/dev/null'
	@echo "Test 7: Get reviews (4000 requests, 150 concurrent)..."
	ab -n 4000 -c 150 "$(API_URL)/api/v1/users/getReview?user_id=u2"
	@echo "Load tests completed!"


