<script>
    import {push} from 'svelte-spa-router'
    import {GetSettings} from "../../wailsjs/go/main/App.js"
    import {shadowConfig} from '../lib/store.js';


    GetSettings().then(result => {
        shadowConfig.set(result);
        push("/welcome")

        // if ("AccessToken" in result) {
        //     const token = result["AccessToken"]
        //
        //     // check if not empty
        //     if (token.length === 0) {
        //         push("/welcome")
        //     }
        //
        //     // check if it can auth
        //     ValidateApiKey(token).then(res => {
        //         if (res === false) {
        //             push("/welcome")
        //         } else {
        //             push("/config")
        //         }
        //     })
        // } else {
        //     push("/welcome")
        // }
    })
</script>

