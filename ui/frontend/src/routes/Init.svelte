<script>
    import {push} from 'svelte-spa-router'
    import {ReadVersion, GetSettings, ValidateApiKey} from "../../wailsjs/go/main/App.js"
    import {shadowConfig, latestVersion, emptyStores} from '../lib/store.js';
    import {isNullOrEmptyObject} from "../lib/utils.js";
    
    async function processVersion() {
        return await ReadVersion().then(result => {
            if (result != null) {
                latestVersion.set(result);
            } else {
                latestVersion.set({})
            }
        }).then(() => {
            return shouldForceUpdate() === true; 
        }) .catch(() => {
            return false
        }) 
    }
    
    async function processAccessToken() {
        if ("AccessToken" in $shadowConfig) {
            const token = $shadowConfig["AccessToken"]
            if (token.length === 0) {
                return '/welcome'
            } else {
                // check if it can auth
                return ValidateApiKey(token).then(res => {
                    if (res !== "") {
                        return '/welcome'
                    } else {
                        return '/config'
                    }
                })
            }
        } else {
            return '/welcome'
        }
    }

    function shouldForceUpdate() {
        return !isNullOrEmptyObject($latestVersion)
            && $latestVersion["should_update"] === true 
            && $latestVersion['current'] !== 'v0.0.0-dev'
            && $latestVersion['download_url'] !== ''
    }

    emptyStores();

    GetSettings().then(result => {
        shadowConfig.set(result);
    }).then(() => {
        processVersion().then(shouldUpdate => {
            if (shouldUpdate === true) {
                push('/update')
            } else {
                processAccessToken().then(page => {
                    push(page)
                })
            }
        })
    })

</script>
