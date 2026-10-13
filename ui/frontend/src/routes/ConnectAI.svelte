<script>
	import './common.css';
	import {onMount} from 'svelte';
	import {push} from 'svelte-spa-router';
	import {GetMCPConnectInfo, WriteMCPConfig} from "../../wailsjs/go/main/App.js";
	import Button from "../components/Button.svelte";
	import StatusBar from "../components/StatusBar.svelte";

	let info = null;
	let loadError = "";
	let writeError = "";
	let justWroteFor = "";
	let copied = "";

	const clients = [
		{ id: "claude", label: "Claude Desktop", hint: "Settings → Developer → Edit Config" },
		{ id: "cursor", label: "Cursor", hint: "~/.cursor/mcp.json" },
		{ id: "claude-code", label: "Claude Code", hint: "run in a terminal" },
		{ id: "other", label: "Other MCP client", hint: "paste the JSON into your client's MCP config" },
	];
	let selectedClient = "claude";

	onMount(async () => {
		try {
			info = await GetMCPConnectInfo();
		} catch (e) {
			loadError = String(e);
		}
	});

	function isConfigured(id) {
		if (!info) return false;
		if (id === "claude") return info.claude_configured;
		if (id === "cursor") return info.cursor_configured;
		return false;
	}

	function canOneClick(id) {
		return id === "claude" || id === "cursor";
	}

	function writeConfig() {
		writeError = "";
		justWroteFor = "";
		WriteMCPConfig(selectedClient).then(() => {
			justWroteFor = selectedClient;
			return GetMCPConnectInfo();
		}).then((updated) => {
			info = updated;
		}).catch((e) => {
			writeError = String(e);
		});
	}

	function copyText(text, tag) {
		navigator.clipboard.writeText(text).then(() => {
			copied = tag;
			setTimeout(() => copied = "", 2000);
		});
	}

	function handleBack() {
		push("/config");
	}
</script>

<div class="connect-ai-page">
	<section class="top-nav">
		<div class="nav-left">
			<div class="block-link" on:click={handleBack} on:keypress={handleBack}>
				<img src="../images/back.svg" alt="back arrow icon">
			</div>
			<div class="h1 p-left-16">Connect an AI assistant</div>
		</div>
	</section>

	<div class="connect-body m-top-8">
		{#if loadError}
			<div class="error-block">{loadError}</div>
		{:else if info}
			<section class="border-weak border-round-8 panel">
				<div class="h3">What can I do with this?</div>
				<div class="text-weak m-top-8">
					Mitti Exporter can run as a local MCP server, so AI assistants can sync and query
					your exported data to answer questions and build reports. For example:
				</div>
				<ul class="examples text-weak">
					<li>"Sync my latest data, then show inspections that failed last month by site."</li>
					<li>"Which actions are overdue, grouped by assignee?"</li>
					<li>"Generate a PDF report for last week's inspections."</li>
					<li>"Build a scorecard for one of my sites."</li>
				</ul>
				<div class="text-weak m-top-8">
					Everything runs locally against a read-only copy of your data. Your API token stays
					in this app — it is never written into any AI client's config.
				</div>

				<div class="status-list m-top-16">
					<div class="status-row">
						<span class="status-dot" class:ok={info.has_token} class:bad={!info.has_token}></span>
						API token configured
					</div>
					<div class="status-row">
						<span class="status-dot" class:ok={info.has_database} class:bad={!info.has_database}></span>
						Local data synced {#if !info.has_database}<span class="text-weak">(run a SQLite export once, or ask the assistant to "sync my data")</span>{/if}
					</div>
				</div>
			</section>

			<section class="border-weak border-round-8 panel m-top-16">
				<div class="h3">Choose your AI client</div>
				<div class="client-picker m-top-8">
					{#each clients as c}
						<div class="client-option" class:selected={selectedClient === c.id} on:click={() => selectedClient = c.id} on:keypress={() => selectedClient = c.id}>
							<div class="client-label">
								{c.label}
								{#if isConfigured(c.id)}<span class="configured-badge">connected</span>{/if}
							</div>
							<div class="text-weak client-hint">{c.hint}</div>
						</div>
					{/each}
				</div>

				{#if selectedClient === "claude-code"}
					<div class="text-weak m-top-16">Run this in a terminal:</div>
					<pre class="snippet m-top-8">{info.claude_code_command}</pre>
					<Button label={copied === "cmd" ? "Copied!" : "Copy command"} type="active-white" onClick={() => copyText(info.claude_code_command, "cmd")}/>
				{:else}
					{#if canOneClick(selectedClient)}
						<div class="m-top-16">
							<Button label={isConfigured(selectedClient) ? "Update config" : "Connect"} type="active-purple" onClick={writeConfig}/>
							{#if justWroteFor === selectedClient && !writeError}
								<div class="text-weak m-top-8">Done — restart {selectedClient === "claude" ? "Claude Desktop" : "Cursor"}, then ask it something like "how many inspections failed last month?".</div>
							{/if}
							{#if writeError}
								<div class="error-block m-top-8">{writeError}</div>
							{/if}
						</div>
					{/if}
					<div class="text-weak m-top-16">
						{#if canOneClick(selectedClient)}Or paste this manually into{:else}Paste this into your client's MCP config ({/if}
						<code>{selectedClient === "claude" ? info.claude_config_path : selectedClient === "cursor" ? info.cursor_config_path : "mcpServers config"}</code>{#if !canOneClick(selectedClient)}){/if}:
					</div>
					<pre class="snippet m-top-8">{info.config_snippet}</pre>
					<Button label={copied === "json" ? "Copied!" : "Copy JSON"} type="active-white" onClick={() => copyText(info.config_snippet, "json")}/>
				{/if}
			</section>
		{:else}
			<div class="text-weak">Loading…</div>
		{/if}
	</div>
</div>

<StatusBar/>

<style>
	.connect-ai-page {
		display: flex;
		flex-direction: column;
		height: 100%;
		padding-bottom: 60px;
	}

	.top-nav {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px 24px 0 24px;
	}

	.nav-left {
		display: flex;
		align-items: center;
	}

	.connect-body {
		padding: 0 24px;
		overflow-y: auto;
	}

	.panel {
		padding: 16px;
	}

	.examples {
		margin: 8px 0 0 0;
		padding-left: 20px;
		line-height: 1.7;
	}

	.status-list {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.status-row {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 0.9rem;
	}

	.status-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
	}

	.status-dot.ok { background-color: #1F8A4C; }
	.status-dot.bad { background-color: #B02010; }

	.client-picker {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}

	.client-option {
		border: 1px solid #D5DBE8;
		border-radius: 8px;
		padding: 10px 12px;
		cursor: pointer;
	}

	.client-option.selected {
		border-color: #6559FF;
		background-color: #F4F3FF;
	}

	.client-label {
		font-weight: 600;
		font-size: 0.9rem;
	}

	.configured-badge {
		background-color: #1F8A4C;
		color: white;
		border-radius: 8px;
		padding: 1px 8px;
		font-size: 0.7rem;
		font-weight: 400;
		margin-left: 6px;
	}

	.client-hint {
		font-size: 0.75rem;
		margin-top: 2px;
	}

	.snippet {
		background-color: #EEF1F7;
		border-radius: 8px;
		padding: 12px;
		font-size: 0.8rem;
		overflow-x: auto;
		white-space: pre;
	}

	.error-block {
		font-size: 0.8rem;
		color: #B02010;
	}
</style>
