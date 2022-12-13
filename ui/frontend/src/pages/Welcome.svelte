<script>
	import './welcome.css';

	import {ValidateApiKey} from "../../wailsjs/go/main/App.js"

	let isValid = false;
	let apiKey;
	let buttonLabel = "Verify"
	let displayBadApiKeyErr = false

	function validate() {
		isValid = false
		ValidateApiKey(apiKey).then((result) => {
			isValid = result
			if (isValid === false) {
				buttonLabel = "Try again"
				displayBadApiKeyErr = true
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
			<h1 class="welcome-title">Welcome to SafetyCulture Exporter</h1>
		</section>
		<section class="token-validation">
			<div class="token-validation-text">Generate an API token from your SafetyCulture <span class="link">user profile</span>.</div>
			<input
				id="token-validation-input"
				type="text"
				placeholder="Enter API Token here"
				bind:value={apiKey}
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

			<button class="token-verify-button" on:click={validate}>{buttonLabel}</button>
		</section>

		<section class="storage-info">
			<div class="note">
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
