package core

const (
	LuaScript = `
		local task_key = KEYS[1];
		if redis.call("SISMEMBER", task_key, ARGV[1]) == 1 then
			return 1;
		end
		redis.call("SADD", task_key, ARGV[1]);
		return 0;
	`
	PkTaskOperator = "task_operator:"
	PkTaskQueue    = "task_queue"
	PkTaskResult   = "task_result:"
	PkTaskFinished = "task_finished:"
)
