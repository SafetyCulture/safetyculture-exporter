<script>
    import Router from 'svelte-spa-router';
    import Init from "./routes/Init.svelte"
    import Welcome from "./routes/Welcome.svelte"
    import Config from "./routes/Config.svelte";
    import Page404 from "./routes/Page404.svelte";
</script>


<Router routes={{
	'/': Init,
	'/welcome': Welcome,
	'/config': Config,
	'*': Page404
}} />
