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
	let justWrote = false;
	let copied = false;

	onMount(async () => {
		try {
			info = await GetMCPConnectInfo();
		} catch (e) {
			loadError = String(e);
		}
	});

	function writeConfig() {
		writeError = "";
		WriteMCPConfig().then(() => {
			justWrote = true;
			return GetMCPConnectInfo();
		}).then((updated) => {
			info = updated;
		}).catch((e) => {
			writeError = String(e);
		});
	}

	function copySnippet() {
		if (!info) return;
		navigator.clipboard.writeText(info.config_snippet).then(() => {
			copied = true;
			setTimeout(() => copied = false, 2000);
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
				<div class="h3">Connect Claude Desktop</div>
				<div class="text-weak m-top-8">
					Mitti Exporter can run as a local MCP server, so Claude can sync and query your
					exported data — inspections, actions, issues, schedules and more — to answer
					questions and build reports. Your API token stays in this app; it is never
					shared with Claude or written into Claude's config.
				</div>

				<div class="status-list m-top-16">
					<div class="status-row">
						<span class="status-dot" class:ok={info.has_token} class:bad={!info.has_token}></span>
						API token configured
					</div>
					<div class="status-row">
						<span class="status-dot" class:ok={info.has_database} class:bad={!info.has_database}></span>
						Local data synced {#if !info.has_database}<span class="text-weak">(run a SQLite export once, or ask Claude to "sync my data")</span>{/if}
					</div>
					<div class="status-row">
						<span class="status-dot" class:ok={info.config_written} class:bad={!info.config_written}></span>
						Claude Desktop config {info.config_written ? "written" : "not yet written"}
					</div>
				</div>

				<div class="m-top-16">
					<Button label={info.config_written ? "Update Claude config" : "Connect to Claude"} type="active-purple" onClick={writeConfig}/>
					{#if justWrote && !writeError}
						<div class="text-weak m-top-8">Done — restart Claude Desktop, then ask it something like "how many inspections failed last month?".</div>
					{/if}
					{#if writeError}
						<div class="error-block m-top-8">{writeError}</div>
					{/if}
				</div>
			</section>

			<section class="border-weak border-round-8 panel m-top-16">
				<div class="h3">Manual setup</div>
				<div class="text-weak m-top-8">
					Prefer to do it yourself, or using another MCP client? Add this to
					<code>{info.host_config_path}</code>:
				</div>
				<pre class="snippet m-top-8">{info.config_snippet}</pre>
				<Button label={copied ? "Copied!" : "Copy"} type="active-white" onClick={copySnippet}/>
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
