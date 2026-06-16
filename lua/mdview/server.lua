local M = {}

local process = nil
local port = nil
local binary = require("mdview.install").binary_path

function M.start(cfg)
	if not vim.uv.fs_stat(binary) then
		vim.notify("[mdview] binary not found at: " .. binary, vim.log.levels.ERROR)
		return
	end

	if process and not process:is_closing() then
		vim.notify("[mdview] server already running", vim.log.levels.WARN)
		return
	end

	local args = { binary }
	table.insert(args, "-host")
	table.insert(args, cfg.host)

	if cfg.port ~= 0 then
		table.insert(args, "-port")
		table.insert(args, tostring(cfg.port))
	end

	if cfg.browser ~= "" then
		table.insert(args, "-browser")
		table.insert(args, cfg.browser)
	end

	process = vim.system(args, {
		text = true,
		detach = true,
		stdout = function(_, data)
			if data and data ~= "" then
				port = tonumber(data:match("PORT:(%d+)"))
				if port then
					vim.schedule(function()
						vim.notify("[mdview] port=" .. port, vim.log.levels.INFO)
					end)
				end
			end
		end,
		stderr = function(_, data)
			if data and data ~= "" then
				vim.schedule(function()
					vim.notify("[mdview] " .. data, vim.log.levels.ERROR)
				end)
			end
		end,
	}, nil)
end

function M.stop()
	if not process or process:is_closing() then
		vim.notify("[mdview] server not running", vim.log.levels.WARN)
	else
		if vim.fn.has("win32") == 1 then
			process:kill("sigint")
		else
			process:kill("sigterm")
		end

		process = nil
		port = nil
		vim.notify("[mdview] server stopped", vim.log.levels.INFO)
	end

	require("mdview.http").close()
end

function M.get_port()
	return port
end

return M
