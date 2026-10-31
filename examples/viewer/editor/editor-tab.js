import { clearHighlights, setEditorDiagnostics } from "./editor.js";
import eventManager from "../ui/event-manager.js";

export let currentTab = "GSL";
export let pendingError = null; // { tab: "GSL" | "WGSL", entries: [...] } or null

export function setTabNames(...names) {
    if (names.length === 1 && Array.isArray(names[0])) {
        names = names[0];
    }
    const container = document.querySelector('.editor-tab-container');
    if (!container) return;
    container.innerHTML = '';

    names.forEach(name => {
        const tabEl = document.createElement('div');
        tabEl.id = `tab-${name}`;
        tabEl.className = 'btn-switch';
        if (name === currentTab) {
            tabEl.classList.add('selected');
        }
        tabEl.textContent = name;
        tabEl.addEventListener('click', () => {
            if (currentTab === name) return;
            selectTab(name);
            clearHighlights();
            refreshPendingError();
        });
        container.appendChild(tabEl);
    });

    if (names.length > 0 && !names.includes(currentTab)) {
        selectTab(names[0]);
    }
}

export function selectTab(tab) {
    if (tab === currentTab) return;
    currentTab = tab;
    const container = document.querySelector('.editor-tab-container');
    if (container) {
        for (const child of container.children) {
            child.classList.toggle('selected', child.textContent === tab);
        }
    }
    eventManager.dispatchEvent("tab/selected", { "tab": tab });
}

export function setPendingError(tab, entries) {
    pendingError = (entries && entries.length) ? { tab, entries } : null;
    const container = document.querySelector('.editor-tab-container');
    if (container) {
        for (const child of container.children) {
            const name = child.textContent;
            child.classList.toggle('has-error', !!pendingError && pendingError.tab === name && currentTab !== name);
        }
    }
    setEditorDiagnostics(pendingError && pendingError.tab === currentTab ? pendingError.entries : []);
}

export function clearPendingError() {
    setPendingError(null, []);
}

export function refreshPendingError() {
    if (pendingError) setPendingError(pendingError.tab, pendingError.entries);
    else clearPendingError();
}
