local M = {}

local defaults = {
	host = "127.0.0.1",
	port = 0,
	browser = "",
	debounce_ms = 50,
	max_mb_body_size = 10,
	render_timeout_sec = 10,
}
local options = {}

function M.setup(opts)
	options = vim.tbl_deep_extend("force", {}, defaults, opts or {})
end

function M.get()
	if vim.tbl_isempty(options) then
		M.setup()
	end
	return vim.deepcopy(options)
end

return M
