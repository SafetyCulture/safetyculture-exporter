<script>
    import './common.css';
    import {CancelExport, ReadExportStatus, OpenDirectory} from "../../wailsjs/go/main/App.js"
    import Status from "./../components/Export/Status.svelte";
    import {shadowConfig, exportConfig} from "../lib/store.js";
    import {onMount} from "svelte";
    import {EventsOn, Quit} from "../../wailsjs/runtime/runtime.js";
    import Button from "../components/Button.svelte";
    import {push} from "svelte-spa-router";
    import Overlay from "../components/Overlay.svelte";
    import StatusBar from "../components/StatusBar.svelte";

    let feedsToExport = $exportConfig['items']
    let exportType = $shadowConfig["Session"]["ExportType"]

    let cancelTriggered = false
    let exportCompleted = false

    onMount(() => {
        EventsOn("finished-export", (newValue) => {
            if (newValue === true) {
                exportCompleted = true
            }
        })
    })

    function handleCancel() {
        cancelTriggered = true
        CancelExport()
    }

    function handleClose() {
        Quit()
    }

    function goBack() {
        push("/config")
    }

    function openExportFolder() {
        OpenDirectory($shadowConfig["Export"]["Path"])
    }

    ReadExportStatus();
</script>

<div class="status-page">
    <section class="top-nav">
        <div class="nav-left">
            <div class="h1">Export status</div>
        </div>
        <div class="nav-right">
            <div class="inline">
                {#if cancelTriggered}
                    <img id="status-cancelled" src='/images/warning-red.svg' alt="export cancelled icon">
                {:else if exportCompleted}
                    <img id="status-completed" src='/images/complete.svg' alt="export completed icon">
                {:else}
                    <img id="status-in-progress" src='/images/in-progress.svg' alt="export in progress icon">
                {/if}
            </div>
            <div class="nav-left inline status-title p-left-8 p-right-16">
                {#if cancelTriggered}
                    Export cancelled
                {:else if exportCompleted}
                    Export complete
                {:else}
                    In progress
                {/if}
            </div>

            {#if !exportCompleted}
                <Button label="Cancel export" type="active-red" onClick={handleCancel}/>
            {:else}
                {#if !cancelTriggered}
                    {#if exportType  === "csv" || exportType === "reports"}
                        <Button label="Open export folder" type="active-white" onClick={openExportFolder}/>
                    {/if}
                    <Button label="Close" clazz="m-left-8" type="active-purple" onClick={handleClose}/>
                {:else}
                    {#if exportType === "reports"}
                        <Button label="Open export folder" type="active-white" onClick={openExportFolder}/>
                    {/if}
                    <Button label="Go Back" type="active-purple" onClick={goBack}/>
                {/if}
            {/if}
        </div>
    </section>

    <div id="overlay-cancel-export">
        {#if cancelTriggered && exportCompleted === false}
            <Overlay>Cancelling export...</Overlay>
        {/if}
    </div>

    <div class="progress-body m-top-16">
        <table class="status-table">
            <thead>
                <tr class="text-weak">
                    <th class="status-col-1">Export item</th>
                    <th class="status-col-2">Status</th>
                    <th class="status-col-3">&nbsp</th>
                </tr>
            </thead>
            <tbody>
            {#each feedsToExport as feed}
                <tr><Status name={feed} cancelled={cancelTriggered}></Status></tr>
            {/each}
            </tbody>
        </table>
    </div>
</div>

<StatusBar/>

<style>
    .status-page {
        padding-top: var(--main-gutter-top);
        padding-left: var(--main-gutter-left);
        padding-right: var(--main-gutter-right);
        background-color: #E9EEF6;
        height: 100%;
    }

    .status-title {
        font-size: 14px;
    }

    .progress-body {
        background-color: white;
        height: calc( 100vh - 200px );
        padding: 20px 16px;
        overflow-y: scroll;
        border-radius: 8px;
    }

    .status-table {
        width: 100%;
        border-collapse: collapse;
    }

    .status-table th {
        font-size: 14px;
        font-weight: 500;
    }

    .status-table tr {
        font-size: 14px;
        border-bottom: 1px solid #EEF1F7;
    }
</style>
