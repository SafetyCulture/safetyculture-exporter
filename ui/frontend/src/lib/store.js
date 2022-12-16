import {writable} from 'svelte/store';

const storedShadowConfig = JSON.parse(localStorage.getItem("cfg"))
export const shadowConfig = writable(storedShadowConfig);
shadowConfig.subscribe(value => {
    console.log("CALLING SUBSCRIBE IN STORE")
    console.log(JSON.stringify(value))
    localStorage.setItem("cfg", JSON.stringify(value));
});
