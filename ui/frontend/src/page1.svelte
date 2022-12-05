<script>
    import { ValidateApiKey } from "../wailsjs/go/main/App.js"

    let isValid = false;
    let apiKey;

    function validate() {
        isValid = false
        ValidateApiKey(apiKey).then((result) => (isValid = result))
    }
</script>

<div>
    <h2>Enter your API key here</h2>
    <input class="input" id="value" type="text" size="60" bind:value={apiKey} />&nbsp;&nbsp;
    {@debug apiKey}
    <button class="btn" on:click={validate}>Validate</button>
</div>

{#if isValid}
<h3>API key is valid</h3>
{:else}
<h3>API key is invalid</h3>
{/if}
