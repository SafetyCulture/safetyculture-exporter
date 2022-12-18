<script>
    import './common.css';
    import {shadowConfig} from "../lib/store.js";

    let data = [
        {
            "id": "1",
            "name": "Template 1",
            "date_modified": "21 Apr 2022"
        },
        {
            "id": "2",
            "name": "Template 2",
            "date_modified": "20 Apr 2022"
        },
        {
            "id": "3",
            "name": "Sample - Electronic Flight Bag (EFB) Assessment for approval Class 1 Software Type A restricted to VFR - duplicate",
            "date_modified": "19 Apr 2022"
        },
    ]

    function trim(org) {
        if (org.length > 80) {
            return org.substring(0, 80).concat(" ...")
        }
        return org
    }

</script>

<div class="template-filter-page p-48">
    <section class="top-nav">
        <div class="nav-left">
            <div class="h1">Export Configuration</div>
        </div>
        <div class="nav-right">
            <button class="button button-white border-round-12">Done</button>
        </div>
    </section>

    <section class="top-nav m-top-16">
        <div class="nav-left">
            <div class="h2">Choose a template</div>
        </div>
        <div class="nav-right">
            <input class="input search" placeholder="Search"/>
        </div>
    </section>

    <section class="m-top-16">
        <div class="table-header text-gray-2">
            <div class="table-row flex-spaced p-horiz-8">
                <div class="nav-left">
                    <input type="checkbox" class="checkbox-purple"/>
                    <div class="m-left-32">Template</div>
                </div>
                <div class="nav-right">
                    <div>Last Modified</div>
                    <div class="m-left-8"><img src="../images/arrow-down.png" alt="down" width="12" height="12"/></div>
                </div>
            </div>
        </div>
        <div class="table-body text-gray-2 m-top-8">
        {#each data as { id, name, date_modified }, i}
            <div class="table-row flex-spaced p-horiz-8">
                <div class="nav-left">
                    <input type="checkbox" class="checkbox-purple" bind:group={$shadowConfig["Export"]["TemplateIds"]} value="{id}"/>
                    <img class="m-left-32" src="../images/template-icon.png" alt="template" width="28" height="28"/>
                    <div class="m-left-8">{trim(name)}</div>
                </div>
                <div class="nav-right">
                    <div>{date_modified}</div>
                </div>
            </div>
        {/each}
        </div>
    </section>
</div>

<style>
    .table-header {
        background-color: #DBDFEB;
    }

    .table-header > .table-row {
        height: 36px;
    }

    .table-body > .table-row {
        height: 52px;
    }

    .table-row {
        display: flex;
        align-items: center;
    }

    input.search {

    }
</style>
