<script>
	import './welcome.css';

	import { ValidateApiKey } from "../../wailsjs/go/main/App.js"
	let isValid = false;
	let apiKey;
	function validate() {
		isValid = false
		ValidateApiKey(apiKey).then((result) => (isValid = result))
	}
</script>

<div class="welcome-page">
	<section class="welcome-left-side">
		<section class="welcome-header">
			<img
				id="welcome-page-logo"
				src="../assets/images/sc-logo.png"
				alt="SafetyCulture logo"
			>
			<p class="welcome-title">Welcome to SafetyCulture Exporter</p>
		</section>
		<section class="token-validation">
			<p class="token-validation-text">
				Generate an API token from your <a href="/">SafetyCulture account.</a>
			</p>
			<input
				id="token-validation-input"
				type="text"
				placeholder="Enter API Token here"
				bind:value={apiKey}
			/>
			<button
				class="token-verify-button"
				on:click={validate}
			>
				Verify
			</button>

			{#if isValid}
				<h3>API key is valid</h3>
			{:else}
				<h3>API key is invalid</h3>
			{/if}
		</section>
		<section class="storage-info">
			<div>
				<img src="" alt="alert icon" width="20" height="20">
			</div>
			<div>
				<p>Important Note</p>
				<p>
					All files (apart from SQL) you export will be stored in the same
					place on your computer or server as the SafetyCulture Exporter.
					If you want to change where your files get exported,
					please move the SafetyCulture Exporter file itself to that place.
				</p>
			</div>
		</section>
	</section>
	<section class="welcome-right-side">
		<div class="image">
			<h2>Image Placeholder</h2>
		</div>
	</section>
</div>
