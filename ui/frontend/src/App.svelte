<script>
    import Router from 'svelte-spa-router';
    import Init from "./routes/Init.svelte"
    import Welcome from "./routes/Welcome.svelte"
    import Config from "./routes/Config.svelte";
    import Page404 from "./routes/Page404.svelte";
    import TemplateFilter from "./routes/TemplateFilter.svelte";
    import FeedFilter from "./routes/FeedFilter.svelte";
    import Debug from "./components/debug.svelte";
</script>

<Debug/>

<Router routes={{
	'/': Init,
	'/welcome': Welcome,
	'/config': Config,
	'/config/templates': TemplateFilter,
	'/config/tables': FeedFilter,
	'*': Page404
}} />


