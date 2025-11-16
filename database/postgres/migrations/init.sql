CREATE TABLE IF NOT EXISTS teams (
    team_name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users (
    user_id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    team_name TEXT REFERENCES teams(team_name),
    is_active BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE,
    merged_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS pr_reviewers (
    id SERIAL PRIMARY KEY,
    pull_request_id TEXT REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id TEXT REFERENCES users(user_id)
);