function ScanUrl(url, params)

end

function textTable(tbl)
	local text = ""
	for k, v in pairs( tbl ) do
	   text = text .. k .. " : " .. v .. "\n"
	end
	return text
end

function GetPathFromUrl(url)
	local path = split(url,"?")[1]
	return path
end

function GetQueryFromUrl(url)
	local query = split(url,"?")
	if #query > 1 then
		return query[2]
	end
	return ""
end

function split(text, sep)
	local startpos = 0
	local endpos = 0
	local t = {}
	local i = 1;

	while startpos ~= nil and endpos ~= nil do
		endpos = string.find(text, sep, startpos+1)
		if endpos == nil then
			endpos = string.len(text)+1
		end
		local param = string.sub(text, startpos+1, endpos-1)
		t[i] = param
		startpos = string.find(text, sep, startpos+1)
		i = i+1
	end
	return t
end

function GetParamsFromQuery(query)
	local k = {}
	if query == nil then
		return k
	end

	local params = split(query,"&")
	for i,v in pairs(params) do
		local kv = split(v,"=")
		if #kv > 1 then
			k[kv[1]] = kv[2]
		else
			k[kv[1]] = ""
		end
	end
	return k
end

function HandleRequest(url, body, method)
	local path = GetPathFromUrl(url)
	local query = GetQueryFromUrl(url)
	local qparams = GetParamsFromQuery(query)

	if path == "/d/settext" then
		gfile.append("test","<h1>ok</h1>")
		return "ok"
	end
	if path == "/d/gettext" then
		local content = gfile.load("test")
		return content
	end
	if path == "/d/queuelua" then
		local idd = gtasks.queueLuaCommand("return 42-23")
		return idd
	end
	if path == "/d/queuenamp" then
		local id = gtasks.queueCommand("nmap")
		return "<b>Task Id: </b>" .. id
	end
	if path == "/d/getqueue" then
		local id = qparams["id"]
		if id == nil then
			return "id not found"
		end

		local output = gtasks.getCommandById(id)
		if output == nil then
			return "id2 not found"
		end

		return "|"..output.."|"
	end

	return "nix"
end