local M = {}
local http = require("mdview.http")
local server = require("mdview.server")

local debounce_timer = assert(vim.uv.new_timer(), "[mdview] failed to create debounce timer")

local function send_content()
	debounce_timer:start(
		100,
		0,
		vim.schedule_wrap(function()
			local bufnr = vim.api.nvim_get_current_buf()
			local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
			local content = table.concat(lines, "\n")
			http.post("/content", { content = content })
		end)
	)
end

local function send_scroll()
	local cursor_line = vim.api.nvim_win_get_cursor(0)[1]
	http.post("/scroll", { cursor_line = cursor_line })
end

local function wait_for_port(callback)
	if server.get_port() then
		callback()
	else
		vim.defer_fn(function()
			wait_for_port(callback)
		end, 30)
	end
end

function M.enable()
	wait_for_port(function()
		send_content()
		send_scroll()
	end)

	local group = vim.api.nvim_create_augroup("mdview", { clear = true })
	vim.api.nvim_create_autocmd({ "TextChanged", "TextChangedI" }, {
		group = group,
		pattern = "*.md",
		callback = send_content,
	})

	vim.api.nvim_create_autocmd({ "CursorMoved", "CursorMovedI" }, {
		group = group,
		pattern = "*.md",
		callback = send_scroll,
	})

	vim.api.nvim_create_autocmd("VimLeavePre", {
		group = group,
		callback = function()
			server.stop()
			M.disable()
		end,
	})
end

function M.disable()
	debounce_timer:stop()
	pcall(vim.api.nvim_del_augroup_by_name, "mdview")
end

return M
