import {writable} from 'svelte/store';

const storedShadowConfig = JSON.parse(localStorage.getItem("cfg"))
export const shadowConfig = writable(storedShadowConfig);
shadowConfig.subscribe(value => {
    localStorage.setItem("cfg", JSON.stringify(value));
});
