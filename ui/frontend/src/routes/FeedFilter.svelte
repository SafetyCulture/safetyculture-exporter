<script>
    import './common.css';
    import {shadowConfig} from "../lib/store.js";

    let data = [
        {
            "left": { "id": "inspections", "name": "Inspections" },
            "right": { "id": "schedules", "name": "Schedules"},
        },
        {
            "left":  { "id": "templates", "name": "Templates" },
            "right": { "id": "template_permissions", "name": "Template Permissions"}
        },
        {
            "left":  { "id": "sites", "name": "Sites"},
            "right": { "id": "site_members", "name": "Site Members"}
        },
        {
            "left":  { "id": "groups", "name": "Groups"},
            "right": { "id": "group_users", "name": "Group Users"}
        },
        {
            "left":  { "id": "schedule_assignees", "name": "Schedule Assignees"},
            "right": { "id": "schedule_occurrences", "name": "Schedule Occurrences"}
        },
        {
            "left":  { "id": "actions", "name": "Actions"},
            "right": { "id": "action_assignees", "name": "Action Assignees"}
        },
        {
            "left":  { "id": "issues", "name": "Issues"},
            "right": { "id": "assets", "name": "Assets"}
        },
        {
            "left":  { "id": "users", "name": "Users"},
            "right": null
        }
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
            <div class="h1">Feed Export</div>
        </div>
        <div class="nav-right">
            <button class="button button-white border-round-12">Done</button>
        </div>
    </section>

    <section class="m-top-16">
        <div class="table-header text-gray-2">
            <div class="table-row p-horiz-8">
                <input type="checkbox" class="checkbox-purple"/>
                <div class="m-left-32">Feed Table</div>
            </div>
        </div>
        <div class="table-body text-gray-2 m-top-8">
        {#each data as { left, right }, i}
            <div class="table-row p-horiz-8">
                {#if left}
                <div class="table-cell">
                    <input type="checkbox" class="checkbox-purple" bind:group={$shadowConfig["Export"]["Tables"]} value="{left.id}"/>
                    <img class="m-left-32" src="../images/template-icon.png" alt="template" width="28" height="28"/>
                    <div class="m-left-8">{trim(left.name)}</div>
                </div>
                {/if}

                {#if right}
                <div class="table-cell">
                    <input type="checkbox" class="checkbox-purple" bind:group={$shadowConfig["Export"]["Tables"]} value="{right.id}"/>
                    <img class="m-left-32" src="../images/template-icon.png" alt="template" width="28" height="28"/>
                    <div class="m-left-8">{trim(right.name)}</div>
                </div>
                {/if}
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
        display: flex;
        align-items: center;
    }

    .table-body > .table-row {
        height: 52px;
    }

    .table-body > .table-row {
        width: 100%;
        display: flex;
    }

    .table-row > .table-cell {
        width: 50%;
        display: flex;
        align-items: center;
    }
</style>
