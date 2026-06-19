local M = {}

local process = nil
local port = nil
local binary = require("mdview.utils").binary_path

function M.start(cfg)
	if not vim.uv.fs_stat(binary) then
		vim.notify(
			"[mdview] binary not found at: " .. binary,
			vim.log.levels.ERROR
		)
		return
	end

	if process and not process:is_closing() then
		vim.notify("[mdview] server already running", vim.log.levels.WARN)
		return
	end

	local args = { binary }
	table.insert(args, "-host")
	table.insert(args, cfg.host)

	table.insert(args, "-port")
	table.insert(args, tostring(cfg.port))

	table.insert(args, "-browser")
	table.insert(args, cfg.browser)

	table.insert(args, "-max-body-size")
	table.insert(args, cfg.max_mb_body_size)

	table.insert(args, "-render-timeout")
	table.insert(args, cfg.render_timeout_sec)

	process = vim.system(args, {
		text = true,
		detach = true,
		stdout = function(_, data)
			if data and data ~= "" then
				port = tonumber(data:match("PORT:(%d+)"))
				if port then
					vim.schedule(function()
						vim.notify(
							"[mdview] port=" .. port,
							vim.log.levels.INFO
						)
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

		process:wait(3000)
		process = nil
		port = nil
		vim.notify("[mdview] server stopped", vim.log.levels.INFO)
	end

	require("mdview.http").close()
end

function M.get_port()
	return port
end

function M.wait_for_ready(timeout_ms, callback)
	if port then
		callback()
		return
	end

	local elapsed = 0
	local timer = vim.uv.new_timer()
	if not timer then
		vim.notify("[mdview] cannot create timer", vim.log.levels.ERROR)
		return
	end

	timer:start(
		0,
		30,
		vim.schedule_wrap(function()
			elapsed = elapsed + 30

			if port then
				timer:stop()
				timer:close()
				callback()
			elseif elapsed >= timeout_ms then
				timer:stop()
				timer:close()
				vim.notify(
					"[mdview] server start timeout",
					vim.log.levels.ERROR
				)
			end
		end)
	)
end

return M
