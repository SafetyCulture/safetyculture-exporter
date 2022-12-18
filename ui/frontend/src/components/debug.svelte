<script>
    import {push} from "svelte-spa-router";

    function gotoWelcome() {
        push("/welcome")
    }

    function gotoConfig() {
        push("/config")
    }

    function gotoTables() {
        push("/config/tables")
    }

    function gotoTemplates() {
        push("/config/templates")
    }
</script>

<div>
    <button on:click={gotoWelcome}>WELCOME</button>
    <button on:click={gotoConfig}>CONFIG</button>
    <button on:click={gotoTemplates}>TEMPLATES</button>
    <button on:click={gotoTables}>TABLES</button>
</div>
