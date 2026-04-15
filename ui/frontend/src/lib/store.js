import {writable} from 'svelte/store';

// shadowConfig will contain YAML configuration shadowed
let storedShadowConfig
try {
    storedShadowConfig = JSON.parse(localStorage.getItem("cfg"))
} catch (e) {
    storedShadowConfig = {}
}
export const shadowConfig = writable(storedShadowConfig);
shadowConfig.subscribe(value => {
    if (value === null) {
        value = {}
    }
    localStorage.setItem("cfg", JSON.stringify(value));
});

// template store
let storedTemplateCache
try {
    storedTemplateCache = JSON.parse(localStorage.getItem("templates"))
} catch (e) {
    storedTemplateCache = []
}
export const templateCache = writable(storedTemplateCache)
templateCache.subscribe(value => {
    if (value === null) {
        value = []
    }
    localStorage.setItem("templates", JSON.stringify(value))
})

// latest version
let storedLatestVersion
try {
    storedLatestVersion = JSON.parse(localStorage.getItem("public-version"))
} catch (e) {
    storedLatestVersion = {}
}
export const latestVersion = writable(storedLatestVersion)
latestVersion.subscribe(value => {
    if (value === null) {
        value = {}
    }
    localStorage.setItem("public-version", JSON.stringify(value))
})

// export status store
let storedExportConfig
try {
    storedExportConfig = JSON.parse(localStorage.getItem("export-config"))
} catch (e) {
    storedExportConfig = {}
}
export const exportConfig = writable(storedExportConfig)
exportConfig.subscribe(value => {
    if (value === null) {
        value = {}
    }
    localStorage.setItem("export-config", JSON.stringify(value))
})

export function emptyStores() {
    templateCache.set([])
    shadowConfig.set({})
    latestVersion.set({})
    exportConfig.set({})
}
