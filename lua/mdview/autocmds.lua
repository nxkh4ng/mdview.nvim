local M = {}
local http = require("mdview.http")
local server = require("mdview.server")
local config = require("mdview.config")

local debounce_timer = assert(vim.uv.new_timer(), "[mdview] failed to create debounce timer")
local prev_line = nil
local prev_content = nil

local function send_content()
	prev_line = nil
	debounce_timer:start(
		config.get().debounce_time,
		0,
		vim.schedule_wrap(function()
			if vim.bo.filetype ~= "markdown" then
				return
			end

			local bufnr = vim.api.nvim_get_current_buf()

			local path = vim.api.nvim_buf_get_name(bufnr)
			local base_dir = vim.fn.fnamemodify(path, ":h")
			base_dir = base_dir:gsub("\\", "/")

			local lines = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
			local content = table.concat(lines, "\n")
			if content == prev_content then
				return
			end
			prev_content = content

			http.post("/content", {
				content = content,
				base_dir = base_dir,
			})
		end)
	)
end

local function send_scroll()
	if vim.bo.filetype ~= "markdown" then
		return
	end

	local cursor_line = vim.api.nvim_win_get_cursor(0)[1]
	if cursor_line == prev_line then
		return
	end
	prev_line = cursor_line
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

	vim.api.nvim_create_autocmd("BufEnter", {
		group = group,
		pattern = "*.md",
		callback = function()
			send_content()
			send_scroll()
		end,
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
