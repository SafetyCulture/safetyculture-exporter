<script>
    export let label = '';
    export let title = ''
    export let onClick
</script>

{#if label !== ''}
    <div class="label">{label}</div>
{/if}
<div class="body" on:click={onClick} on:keypress={onClick}>
    <div>{title}</div>
    <div class="button-right">
        <img src="../images/arrow-right-compact.svg" alt="right arrow icon">
    </div>
</div>


<style>
    .label {
        margin-top: 15px;
        font-weight: 500;
        font-size: 0.9rem;
        line-height: 1rem;
    }

    .body {
        display: flex;
        justify-content: space-between;

        margin-top: 4px;
        padding: 12px 11px 12px 16px;
        font-weight: 400;
        font-size: 1rem;
        line-height: 1rem;

        border: 1px solid #BFC5D4;
        border-radius: 8px;

        cursor: pointer;
    }

    .button-right {
        display: flex;
        align-items: center;
    }
</style>
