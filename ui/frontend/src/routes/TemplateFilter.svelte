<script>
    import './common.css';
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

    <section class="top-nav">
        <div class="nav-left">
            <div class="h2">Choose a template</div>
        </div>
        <div class="nav-right">
            <input class="input search" placeholder="Search"/>
        </div>
    </section>

    <section>
        <div class="table-header">
            <div class="table-row top-nav">
                <div class="nav-left">
                    <input type="checkbox"/>
                    <div>Template</div>
                </div>
                <div class="nav-right">
                    <div>Last Modified</div>
                    <div><img src="../images/arrow-down.png" alt="down" width="12" height="12"></div>
                </div>
            </div>
        </div>

    </section>
</div>

<style>
    .top-nav {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .nav-left {
        display: flex;
        align-items: center;
        justify-content: flex-start;
    }

    .nav-right {
        display: flex;
        align-items: center;
        justify-content: flex-end;
    }

    .table-header {
        background-color: #DBDFEB;
    }

    input.search {

    }
</style>
