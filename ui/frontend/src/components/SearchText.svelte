<script>
    export let clazz = '';
    export let placeholder = ''
    export let value = '';

    function emptyFn() {
        value = '';
    }
</script>

<div class="input-container">
    <img src="../images/search.svg" alt="&nbsp;"/>
    <input class="input p-left-8 p-right-8 {clazz}" placeholder={placeholder} bind:value={value}/>
    <img src="../images/close-circle.svg" class="pointer" alt="&nbsp;" on:click={emptyFn} on:keydown={emptyFn}/>
</div>


<style>
    .input-container {
        display: flex;
        width: 100%;
        padding: 10px 10px;
        border-radius: 8px;
        border: 1px solid #BFC5D4;
        font-size: 1rem;
        line-height: 1.5rem;
    }

    .input {
        border: none;
        outline: none;
        font-family: NotoSansRegular, Tahoma, sans-serif;
        font-size: 14px;
    }

    img.pointer {
        cursor: pointer;
    }
</style>
