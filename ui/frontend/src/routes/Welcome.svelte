<script>
	import './common.css';

	import {push} from 'svelte-spa-router'
	import {ValidateApiKey} from "../../wailsjs/go/main/App.js"
	import {shadowConfig} from '../lib/store.js';


	let isValid = false;
	let buttonLabel = "Verify"
	let displayBadApiKeyErr = false

	function validate() {
		isValid = false
		ValidateApiKey($shadowConfig["AccessToken"]).then((result) => {
			isValid = result
			if (isValid === false) {
				buttonLabel = "Try again"
				displayBadApiKeyErr = true
			} else {
				push("/config")
			}
		})
	}
</script>
<div class="welcome-page">
	<section class="welcome-left-side">
		<section class="welcome-header">
			<img
				id="welcome-page-logo"
				src="../images/logo.png"
				alt="SafetyCulture logo"
			/>
			<div class="h1">Welcome to SafetyCulture Exporter</div>
		</section>
		<section class="token-validation">
			<div class="token-validation-text">Generate an API token from your SafetyCulture <span class="link">user profile</span>.</div>
			<input
				class="input"
				type="text"
				placeholder="Enter API Token here"
				bind:value={$shadowConfig["AccessToken"]}
			/>

			{#if displayBadApiKeyErr}
				<div class="error-block">
					<div class="error-block-title">
						Oops! We couldn't verify your token.
					</div>
					<div class="error-block-body">
						Here's our diagnosis:
						<ul>
							<li>Your token might be expired. It expires after 30 days of not being used.</li>
						</ul>
					</div>
				</div>
			{/if}

			<button class="button button-purple m-top-8 border-round-12" on:click={validate}>{buttonLabel}</button>
		</section>

		<section class="storage-info">
			<div class="note border-round-8">
				<div>
					<img src="../images/round_exclamation_mark.png" alt="alert icon" width="20" height="20">
				</div>
				<div>
					<div class="note-title">Before you continue</div>
					<div class="note-body">All files (apart from SQL) you export will be stored in the same place on your computer or server as the SafetyCulture Exporter. If you want to change where your files get exported, please move the SafetyCulture Exporter file itself to that place.</div>
				</div>
			</div>
		</section>
	</section>
	<section class="welcome-right-side">
		<div class="right-image">
			<img src="../images/token_example.png" alt="example generating token">
		</div>
	</section>
</div>

<style>
	.welcome-page {
		display: flex;
	}

	.welcome-left-side {
		width: 50%;
		padding: 1.5rem;
	}

	.welcome-right-side {
		width: 50%;
	}

	#welcome-page-logo {
		width: 150px;
	}

	div.token-validation-text {
		margin-bottom: 8px;
		color: #1D2330;
	}

	.token-validation-text {
		font-style: normal;
		font-weight: 400;
		font-size: 1rem;
		line-height: 1.5rem;
	}

	.note {
		background-color: #EEF1F7;
		padding: 16px;
		color: #3F495A;
		display: flex;
		flex-direction: row;
		gap: 16px;
		font-size: 0.9rem;
	}

	.note .note-title {
		font-weight: bold;
		padding-bottom: 8px;
	}

	.note .note-body {
		line-height: 1.5rem;
	}

	div.right-image {
		max-width: 100%;
		max-height: 100%;
	}

	.right-image img {
		height: 100%;
		width: 100%;
		object-fit: contain;
	}

	div.error-block {
		font-size: 0.8rem;
		color: #1D2330;
	}

	div.error-block .error-block-title {
		color: #A02228;
	}

	div.error-block .error-block-body ul {
		margin-top: 2px;
		margin-bottom: 2px;
	}

	section.storage-info {
		margin-top: 24px;
	}
</style>
