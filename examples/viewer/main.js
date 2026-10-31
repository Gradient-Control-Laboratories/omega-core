import { showNotification, clearNotification } from "./ui/notification.js";
import { clearError, showError, startErrorTimer, stopErrorTimer, resetErrorTimer } from "./ui/error.js";
import { ProgressBar, operationKindLabels } from "./ui/progress-bar.js";
import eventManager from "./ui/event-manager.js";
import "./ui/panel-resize.js";
import "./ui/tooltip.js";
import {
    updateShader,
    setParts,
    setShapeColor,
    setSelectedShape,
    setFieldEnabled,
    setFieldHeight,
    setFieldNormal,
    setFieldInterval,
    setFieldFade,
    setModelOpacity,
    setSectionOpacity,
    setFieldUnion,
    setFieldExternal,
    setShapeVisible,
    setAnimPaused,
    setAnimTime,
    getAnimTime,
    audioSystem,
    pickWorldPoint,
    pickWorldPointNormal,
    addMesh,
    removeMesh,
    getMeshEntries,
    getMeshGeometry,
    syncMeshInstances,
    setOutlinesEnabled,
    setWorldExtent,
    setGridDensity,
    frameWorldExtent,
    getWorldExtent,
    getGridDensity,
    setSliderValue,
    setUiControlValues,
} from "./render/viewer.js";
import { srgbEncode } from "./render/gsl-shading.js";
import {
    initEditor,
    getContent,
    setContent,
    highlightSpans,
    clearHighlights,
    insertAtCursor,
    setLineWrap,
    replaceSourceSpan,
} from "./editor/editor.js";
import {
    currentTab,
    selectTab,
    setPendingError,
    clearPendingError,
    setTabNames,
} from "./editor/editor-tab.js";
import { initHelpPanel } from "./docs/help-panel.js";
import { initSliderPanel, renderSliderPanel, sliderAnimTick } from "./panels/slider-panel.js";
import { initMeshImport } from "./panels/mesh-import.js";
import { initExportPanel } from "./panels/export-panel.js";

const clientId = (typeof crypto !== "undefined" && crypto.randomUUID)
    ? crypto.randomUUID()
    : `client-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

const progressBar = new ProgressBar({ targetContainer: "#canvas-container" });

let compilePending = false;
let compileSeq = 0;
let latestCompileID = "";
let debounceTimer = null;
let currentGSL = "";
let currentWGSL = "";
let gslSource = "";
let previousTab = "GSL";
let currentShapeLocations = [];
let currentExampleName = "";
let suppressDirty = false;
let isDirty = false;

const btnToggleCode = document.getElementById("btn-toggle-code");
const editorContainerEl = document.getElementById("editor-container");
const splitterEl = document.querySelector(".splitter");
const editorTopControlsEl = document.getElementById("editor-top-controls");
const exampleSelectEl = document.getElementById("example-select");
const btnSaveAs = document.getElementById("btn-save-as");
const saveAsPopover = document.getElementById("save-as-popover");
const saveAsInput = document.getElementById("save-as-input");
const btnSaveConfirm = document.getElementById("btn-save-confirm");
const btnSaveCancel = document.getElementById("btn-save-cancel");

function nextCompileID() {
    latestCompileID = `compile-${Date.now()}-${++compileSeq}`;
    return latestCompileID;
}

function markCompileStarted() { compilePending = true; }
function markCompileResolved() { compilePending = false; }

function setDirty(dirty) {
    isDirty = dirty;
    btnSaveAs.disabled = !dirty;
}

function markDirty() {
    if (!suppressDirty) {
        setDirty(true);
    }
}

function defaultSaveCopyName(name) {
    const base = (name || "example.gsl").replace(/\.gsl$/i, "");
    return `${base}_copy.gsl`;
}

// ── Top Bar: Code Slide-Out Toggle ──

function setCodePanelOpen(open) {
    editorContainerEl.classList.toggle("collapsed", !open);
    splitterEl.classList.toggle("hidden", !open);
    editorTopControlsEl.classList.toggle("hidden", !open);
    btnToggleCode.classList.toggle("active", open);
    window.dispatchEvent(new Event("resize"));
}

btnToggleCode.addEventListener("click", () => {
    const isCollapsed = editorContainerEl.classList.contains("collapsed");
    setCodePanelOpen(isCollapsed);
});

// ── Top Bar: Save As Popover ──

function closeSaveAsPopover() {
    saveAsPopover.classList.add("hidden");
}

btnSaveAs.addEventListener("click", (e) => {
    e.stopPropagation();
    if (btnSaveAs.disabled) return;
    const opening = saveAsPopover.classList.contains("hidden");
    if (opening) {
        saveAsInput.value = defaultSaveCopyName(currentExampleName);
        saveAsPopover.classList.remove("hidden");
        saveAsInput.focus();
        saveAsInput.select();
    } else {
        closeSaveAsPopover();
    }
});

async function submitSaveAs() {
    let name = (saveAsInput.value || "").trim();
    if (!name) return;
    if (!name.toLowerCase().endsWith(".gsl")) {
        name += ".gsl";
    }
    const source = getGslSourceForServer();
    try {
        const res = await fetch("/api/save", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name, source }),
        });
        const data = await res.json();
        if (!res.ok || data.error) {
            throw new Error(data.error || `HTTP ${res.status}`);
        }
        const savedName = data.name || name;
        currentExampleName = savedName;
        await refreshExampleList(savedName);
        setDirty(false);
        closeSaveAsPopover();
        showNotification(false, `Saved ${savedName}`);
    } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        showNotification(true, `Save failed: ${msg}`);
    }
}

btnSaveConfirm.addEventListener("click", (e) => {
    e.stopPropagation();
    submitSaveAs();
});

if (btnSaveCancel) {
    btnSaveCancel.addEventListener("click", (e) => {
        e.stopPropagation();
        closeSaveAsPopover();
    });
}

saveAsInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        e.preventDefault();
        submitSaveAs();
    } else if (e.key === "Escape") {
        e.preventDefault();
        closeSaveAsPopover();
    }
});

document.addEventListener("click", (e) => {
    if (!saveAsPopover.classList.contains("hidden") &&
        !saveAsPopover.contains(e.target) &&
        e.target !== btnSaveAs) {
        closeSaveAsPopover();
    }
});

// ── Error Formatting & Diagnostics ──

function formatErrBody(e) {
    const parts = [];
    if (e.line) {
        parts.push(e.line);
        if (e.column >= 1 && e.length >= 1) {
            parts.push(" ".repeat(e.column - 1) + "^".repeat(e.length) + " " + e.message);
        } else {
            parts.push(e.message);
        }
    } else {
        parts.push(e.message);
    }
    return parts.join("\n");
}

function gslPlainTextErrorToEntry(text) {
    const trimmed = (typeof text === "string" ? text : String(text || "")).trim();
    const caretMatch = trimmed.match(/\n\s*\^\s*(\d+):(\d+):\s*([^\n]*)/)
        || trimmed.match(/^\s*\^\s*(\d+):(\d+):\s*(.*)$/m);
    if (caretMatch) {
        return {
            line: +caretMatch[1],
            col: +caretMatch[2],
            length: 1,
            message: trimmed,
            actions: [],
        };
    }
    return {
        line: 1, col: 1, length: 1,
        message: trimmed,
        actions: [],
    };
}

eventManager.addListener("canvas/errors", (payload) => {
    markCompileResolved();
    stopErrorTimer();
    progressBar.reset();

    let entries = [];
    let joined = "";
    if (typeof payload === "string") {
        joined = payload;
        entries = [{ line: 1, col: 1, length: 1, message: payload, actions: [] }];
    } else if (payload && Array.isArray(payload.errors)) {
        for (const e of payload.errors) {
            const body = formatErrBody(e);
            entries.push({
                line: e.lineNum || 1,
                col: e.column || 1,
                length: Math.max(1, e.length || 1),
                message: body,
                actions: [],
            });
        }
        joined = payload.errors.map(formatErrBody).join("\n");
    } else if (Array.isArray(payload)) {
        const chunks = [];
        for (const item of payload) {
            const s = typeof item === "string" ? item : "";
            if (s.trim() === "") continue;
            chunks.push(s.trim());
            entries.push(gslPlainTextErrorToEntry(s.trim()));
        }
        joined = chunks.join("\n\n");
    }

    setPendingError("GSL", entries);
    showError(joined, false);
});

eventManager.addListener("error", (event) => {
    if (!event || !event.message) {
        clearPendingError();
        return;
    }
    if (!event.wgsl) return;
    const probably = event.message.match(/probably line (\d+)/);
    const line = probably ? +probably[1] : 1;
    setPendingError("WGSL", [{
        line, col: 1, length: 1,
        message: event.message,
        actions: [],
    }]);
});

// ── 3D Gizmos & Shape Picking ──

eventManager.addListener("gsl-ui/commit", (event) => {
    if (!event.span || typeof event.text !== "string") return;
    markDirty();
    replaceSourceSpan(event.span, event.text);
});

eventManager.addListener("gsl-ui/input", (event) => {
    setUiControlValues(event);
});

eventManager.addListener("shape/picked", (event) => {
    const shapeIndex = event.shapeIndex;
    setSelectedShape(shapeIndex);

    if (shapeIndex > 0 && !colorPickerPanel.classList.contains("open")) {
        colorPickerPanel.classList.add("open");
    }

    colorPickersContainer.querySelectorAll(".color-picker-row").forEach((row) => {
        if (parseInt(row.dataset.shapeIndex, 10) === shapeIndex) {
            row.classList.add("selected");
        } else {
            row.classList.remove("selected");
        }
    });

    if (shapeIndex === 0) {
        clearHighlights();
        return;
    }

    const locations = currentShapeLocations[shapeIndex - 1];
    if (locations && locations.spans && currentTab === "GSL") {
        highlightSpans(locations.spans);
    }
});

function detectShapeCount(wgslCode) {
    const matches = wgslCode.matchAll(/gsl_shape\([^,]+,\s*(\d+)u\)/g);
    let maxIndex = 0;
    for (const match of matches) {
        maxIndex = Math.max(maxIndex, parseInt(match[1], 10));
    }
    return Math.max(maxIndex, 1);
}

eventManager.addListener("canvas/update-shader", (event) => {
    if (event.wgsl) {
        currentWGSL = event.wgsl;
        currentShapeLocations = event.shapeLocations || [];
        setParts(event.parts || []);
        updateShader(
            event.wgsl,
            event.shapeCount || 1,
            event.shapeNames || [],
            event.meshSdfSlots || event.meshSdfBindings || [],
            event.sliders || [],
            event.uiControls || [],
        );
        markCompileResolved();
        clearPendingError();
        clearError();
        stopErrorTimer();
        progressBar.reset();
        clearHighlights();
        setSelectedShape(0);

        const targetExtent = (event.scene && event.scene.size > 0) ? event.scene.size / 2 : 4;
        const sizeChanged = Math.abs(targetExtent - getWorldExtent()) > 1e-4;
        setWorldExtent(targetExtent);
        if (event.scene && event.scene.grid > 0) {
            setGridDensity(event.scene.grid);
        }
        if (sizeChanged) frameWorldExtent();

        const decls = event.meshDecls || [];
        const unknowns = syncMeshInstances(decls);
        if (unknowns.length > 0) {
            showError(`Unknown mesh index: ${unknowns.join(", ")}. Upload in the Meshes panel first.`, false);
        }

        renderSliderPanel(event.sliders || [], event.uiControls || []);

        if (currentTab === "WGSL") {
            suppressDirty = true;
            setContent(currentWGSL);
            suppressDirty = false;
        }
    }
});

// ── HTTP /api/compile & /api/mesh Client ──

async function compileGsl(source) {
    if (!source || source.trim() === "") return;
    const reqId = nextCompileID();
    markCompileStarted();
    startErrorTimer("compiler");
    showNotification(false, "Compiling...");

    try {
        const res = await fetch("/api/compile", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                id: reqId,
                clientId,
                source,
            }),
        });
        const data = await res.json();
        if (reqId !== latestCompileID) return;

        if (!res.ok || (data.errors && data.errors.length > 0)) {
            const errs = data.errors && data.errors.length > 0
                ? data.errors
                : [`Compile failed (HTTP ${res.status})`];
            eventManager.dispatchEvent("canvas/errors", errs);
            return;
        }

        clearNotification();
        eventManager.dispatchEvent("canvas/update-shader", data);

        const audioJs = data.audio && (data.audio.jsCode || data.audio.javascript) || "";
        const hasAudio = Boolean(audioJs);
        eventManager.dispatchEvent("canvas/update-audio", {
            javascript: audioJs,
            hasAudio,
        });
    } catch (err) {
        if (reqId !== latestCompileID) return;
        const msg = err instanceof Error ? err.message : String(err);
        eventManager.dispatchEvent("canvas/errors", [msg]);
    }
}

async function requestStlExport(detail, surfaceBbox) {
    const source = getGslSourceForServer();
    if (!source || source.trim() === "") {
        showNotification(true, "No code to export");
        return false;
    }

    const reqId = `mesh-${Date.now()}`;
    const subdivision = detail ?? 7;
    const ext = getWorldExtent();
    const bbox = Array.isArray(surfaceBbox) && surfaceBbox.length === 6
        ? surfaceBbox
        : [-ext, -ext, -ext, ext, ext, ext];

    showNotification(false, "Generating STL mesh...", true);
    startErrorTimer("stl export");

    try {
        const res = await fetch("/api/mesh", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                id: reqId,
                clientId,
                source,
                subdivision,
                bbox,
                algo: "marching-cubes",
                time: getAnimTime(),
            }),
        });

        stopErrorTimer();
        progressBar.reset();

        if (!res.ok) {
            let errText = `HTTP ${res.status}`;
            try {
                const errJson = await res.json();
                if (errJson && errJson.error) errText = errJson.error;
            } catch (_) {
                // ignore JSON parse error on non-JSON body
            }
            showError(`STL export failed: ${errText}`, false, true);
            return false;
        }

        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        const baseName = (currentExampleName || "model.gsl").replace(/\.gsl$/i, "");
        a.href = url;
        a.download = `${baseName}.stl`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        setTimeout(() => URL.revokeObjectURL(url), 1000);
        showNotification(false, `Exported ${baseName}.stl`);
        return true;
    } catch (err) {
        stopErrorTimer();
        progressBar.reset();
        const msg = err instanceof Error ? err.message : String(err);
        showError(`STL export failed: ${msg}`, false, true);
        return false;
    }
}

// ── Streaming NDJSON Progress Reader (/api/progress) ──

function handleProgressEvent(ev) {
    if (!ev) return;
    resetErrorTimer();
    const stage = ev.stage || "compile";
    if (ev.done) {
        progressBar.reset();
        return;
    }
    const label = ev.label || operationKindLabels[stage] || stage;
    progressBar.startProgress(stage, label, 10000, ev.completed || 0, ev.total || 0);
}

async function connectProgressStream() {
    while (true) {
        try {
            const res = await fetch(`/api/progress?clientId=${encodeURIComponent(clientId)}`);
            if (!res.ok || !res.body) {
                throw new Error(`HTTP ${res.status}`);
            }
            const reader = res.body.getReader();
            const decoder = new TextDecoder();
            let buffer = "";
            while (true) {
                const { value, done } = await reader.read();
                if (done) break;
                buffer += decoder.decode(value, { stream: true });
                let newlineIdx;
                while ((newlineIdx = buffer.indexOf("\n")) >= 0) {
                    const line = buffer.slice(0, newlineIdx).trim();
                    buffer = buffer.slice(newlineIdx + 1);
                    if (!line) continue;
                    try {
                        handleProgressEvent(JSON.parse(line));
                    } catch (parseErr) {
                        console.warn("Invalid progress JSON line:", parseErr);
                    }
                }
            }
        } catch (err) {
            // Reconnect after a short delay if the server restarts or closes the stream.
            await new Promise((r) => setTimeout(r, 1500));
        }
    }
}

connectProgressStream();

// ── Surface Point & Point+Normal Picking Tools ──

const btnPointSelect = document.getElementById("btn-point-select");
const btnPointNormalSelect = document.getElementById("btn-point-normal-select");
let pointSelectMode = false;
let pointNormalSelectMode = false;

function topoToolActive() {
    return pointSelectMode || pointNormalSelectMode;
}

function enterTopoToolVisuals(msg) {
    document.getElementById("canvas-container").style.cursor = "crosshair";
    setOutlinesEnabled(false);
    showNotification(false, msg);
}

function exitTopoToolVisuals() {
    document.getElementById("canvas-container").style.cursor = "";
    setOutlinesEnabled(true);
    clearNotification();
}

function exitPointSelectMode() {
    pointSelectMode = false;
    btnPointSelect.classList.remove("active");
    exitTopoToolVisuals();
}

function enterPointSelectMode() {
    if (pointNormalSelectMode) exitPointNormalSelectMode();
    pointSelectMode = true;
    btnPointSelect.classList.add("active");
    enterTopoToolVisuals("Click on a surface to insert [x, y, z]");
}

function exitPointNormalSelectMode() {
    pointNormalSelectMode = false;
    btnPointNormalSelect.classList.remove("active");
    exitTopoToolVisuals();
}

function enterPointNormalSelectMode() {
    if (pointSelectMode) exitPointSelectMode();
    pointNormalSelectMode = true;
    btnPointNormalSelect.classList.add("active");
    enterTopoToolVisuals("Click on a surface to insert a point + normal");
}

btnPointSelect.addEventListener("click", () => {
    if (pointSelectMode) exitPointSelectMode();
    else enterPointSelectMode();
});

btnPointNormalSelect.addEventListener("click", () => {
    if (pointNormalSelectMode) exitPointNormalSelectMode();
    else enterPointNormalSelectMode();
});

document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
        if (pointSelectMode) exitPointSelectMode();
        else if (pointNormalSelectMode) exitPointNormalSelectMode();
    }
});

let canvasMouseDownPos = null;
document.getElementById("canvas-container").addEventListener("mousedown", (event) => {
    if (document.getElementById("canvas-controls").contains(event.target)) return;
    if (document.getElementById("canvas-left-btns").contains(event.target)) return;
    canvasMouseDownPos = { x: event.clientX, y: event.clientY };
});

document.getElementById("canvas-container").addEventListener("click", async (event) => {
    if (!topoToolActive()) return;
    if (btnPointSelect.contains(event.target) || btnPointNormalSelect.contains(event.target)) return;
    if (document.getElementById("canvas-controls").contains(event.target)) return;
    if (document.getElementById("canvas-left-btns").contains(event.target)) return;

    if (canvasMouseDownPos) {
        const dx = event.clientX - canvasMouseDownPos.x;
        const dy = event.clientY - canvasMouseDownPos.y;
        if (Math.sqrt(dx * dx + dy * dy) > 3) {
            canvasMouseDownPos = null;
            return;
        }
    }
    canvasMouseDownPos = null;

    event.stopPropagation();
    event.preventDefault();

    if (pointNormalSelectMode) {
        const hit = await pickWorldPointNormal(event);
        if (hit) {
            const p = `[${hit.point.x.toFixed(2)}, ${hit.point.y.toFixed(2)}, ${hit.point.z.toFixed(2)}]`;
            const n = hit.normal
                ? `[${hit.normal.x.toFixed(4)}, ${hit.normal.y.toFixed(4)}, ${hit.normal.z.toFixed(4)}]`
                : `[0, 0, 1]`;
            markDirty();
            insertAtCursor(`${p}, ${n}`);
        } else {
            showNotification(true, "No surface hit — try clicking on a shape or mesh");
        }
        return;
    }

    const point = await pickWorldPoint(event);
    if (point) {
        const text = `[${point.x.toFixed(2)}, ${point.y.toFixed(2)}, ${point.z.toFixed(2)}]`;
        markDirty();
        insertAtCursor(text);
    } else {
        showNotification(true, "No surface hit — try clicking on a shape or mesh");
    }
}, true);

const btnFullscreen = document.getElementById("btn-fullscreen");
btnFullscreen.addEventListener("click", () => {
    const container = document.getElementById("canvas-container");
    if (!document.fullscreenElement) {
        container.requestFullscreen();
    } else {
        document.exitFullscreen();
    }
});

// ── CodeMirror Editor & GSL/WGSL Tabs ──

const editorHost = document.getElementById("editor");
const initialContent = `// Select an example from the dropdown above\n// or write your own GSL code here\n`;

function getGslSourceForServer() {
    if (currentTab === "GSL") {
        return getContent();
    }
    return gslSource || currentGSL || "";
}

function onEditorChange(content) {
    if (suppressDirty) return;
    markDirty();
    if (debounceTimer) {
        clearTimeout(debounceTimer);
    }
    debounceTimer = setTimeout(() => {
        clearHighlights();
        if (currentTab === "GSL") {
            gslSource = content;
            currentGSL = content;
            compileGsl(content);
        } else if (currentTab === "WGSL") {
            currentWGSL = content;
            eventManager.dispatchEvent("canvas/update-shader", {
                wgsl: content,
                shapeCount: detectShapeCount(content),
            });
            clearError();
            stopErrorTimer();
        }
    }, 500);
}

initEditor(editorHost, initialContent.trim(), onEditorChange);

eventManager.addListener("tab/selected", (event) => {
    const nextTab = event.tab;
    if (previousTab === "GSL") {
        gslSource = getContent();
    } else if (previousTab === "WGSL") {
        currentWGSL = getContent();
    }

    suppressDirty = true;
    if (nextTab === "GSL") {
        setContent(gslSource);
    } else if (nextTab === "WGSL") {
        setContent(currentWGSL);
    }
    suppressDirty = false;
    previousTab = nextTab;
});

setTabNames("GSL", "WGSL");

let wrapEnabled = true;
const btnWrap = document.getElementById("btn-wrap");
btnWrap.addEventListener("click", () => {
    wrapEnabled = !wrapEnabled;
    btnWrap.classList.toggle("active", wrapEnabled);
    setLineWrap(wrapEnabled);
});

// ── Help Panel, Slider Panel, Mesh Import Panel, Export Panel ──

initHelpPanel({
    menuHost: document.getElementById("help-menu"),
    panelHost: document.getElementById("canvas-container"),
});

initSliderPanel({
    setSliderValue,
    getAnimTime,
    audioSystem,
    onCommit: () => {
        markDirty();
        compilePending = true;
    },
});

document.getElementById("slider-panel-header").addEventListener("click", (e) => {
    const panel = document.getElementById("slider-panel");
    if (e.target.classList.contains("panel-close")) {
        panel.classList.remove("open");
    } else if (!panel.classList.contains("open")) {
        panel.classList.add("open");
    }
});

initMeshImport({
    addMesh,
    removeMesh,
    getMeshEntries,
    getSource: getGslSourceForServer,
    setSource: (next) => {
        if (currentTab !== "GSL") selectTab("GSL");
        suppressDirty = true;
        setContent(next);
        suppressDirty = false;
        gslSource = next;
        currentGSL = next;
        markDirty();
    },
    recompile: () => {
        const source = getGslSourceForServer();
        if (!source || source.trim() === "") return;
        compileGsl(source);
    },
    dropTarget: document.getElementById("canvas-container"),
});

initExportPanel({
    getMeshGeometry,
    getWorldExtent,
    getSource: getGslSourceForServer,
    clearError,
    onServerExport: (detail, surfaceBbox) => {
        return requestStlExport(detail, surfaceBbox);
    },
});

// ── Palette & Shapes Panel ──

const PALETTE_PRESETS = [
    { a: [0.5, 0.5, 0.5], b: [0.5, 0.5, 0.5], c: [1, 1, 1], d: [0.00, 0.33, 0.67] },
    { a: [0.5, 0.5, 0.5], b: [0.55, 0.55, 0.55], c: [1, 1, 1], d: [0.30, 0.50, 0.50] },
    { a: [0.5, 0.5, 0.5], b: [0.5, 0.5, 0.5], c: [1, 1, 1], d: [0.00, 0.10, 0.20] },
    { a: [0.5, 0.5, 0.5], b: [0.5, 0.5, 0.5], c: [1, 1, 1], d: [0.30, 0.20, 0.20] },
    { a: [0.5, 0.5, 0.5], b: [0.5, 0.5, 0.5], c: [1, 1, 0.5], d: [0.80, 0.90, 0.30] },
    { a: [0.5, 0.5, 0.5], b: [0.5, 0.5, 0.5], c: [1, 0.7, 0.4], d: [0.00, 0.15, 0.20] },
];

let currentPalette = PALETTE_PRESETS[0];
let paletteShapeCount = 0;

function cosineRGB(t, { a, b, c, d }) {
    return a.map((ai, i) => Math.max(0, Math.min(1, ai + b[i] * Math.cos(2 * Math.PI * (c[i] * t + d[i]))))).map(srgbEncode);
}

function rgbToHex([r, g, b]) {
    const h = (v) => Math.round(v * 255).toString(16).padStart(2, "0");
    return `#${h(r)}${h(g)}${h(b)}`;
}

function shapeT(i) { return (1 / 3 + (i - 1) * (5 / 12)) % 1; }

const colorPickerPanel = document.getElementById("color-picker-panel");
const colorPickersContainer = document.getElementById("color-pickers");
const paletteStrip = document.getElementById("palette-strip");

document.getElementById("color-picker-header").addEventListener("click", (e) => {
    if (e.target.classList.contains("panel-close")) {
        colorPickerPanel.classList.remove("open");
    } else if (!colorPickerPanel.classList.contains("open")) {
        colorPickerPanel.classList.add("open");
    }
});

function renderPaletteStrip(palette) {
    paletteStrip.innerHTML = "";
    for (let k = 0; k < 12; k++) {
        const [r, g, b] = cosineRGB(k / 12, palette);
        const el = document.createElement("div");
        el.className = "palette-swatch";
        el.style.background = `rgb(${Math.round(r * 255)},${Math.round(g * 255)},${Math.round(b * 255)})`;
        paletteStrip.appendChild(el);
    }
}

function applyPalette(palette) {
    currentPalette = palette;
    renderPaletteStrip(palette);
    for (let i = 1; i <= paletteShapeCount; i++) {
        const [r, g, b] = cosineRGB(shapeT(i), palette);
        setShapeColor(i, r, g, b);
        const row = colorPickersContainer.querySelector(`.color-picker-row[data-shape-index="${i}"]`);
        if (row) row.querySelector('input[type="color"]').value = rgbToHex([r, g, b]);
    }
    updatePopupHeader();
}

renderPaletteStrip(currentPalette);

function drawPresetCanvas(canvas, palette) {
    const ctx = canvas.getContext("2d");
    const w = canvas.width, h = canvas.height;
    for (let x = 0; x < w; x++) {
        const [r, g, b] = cosineRGB(x / w, palette);
        ctx.fillStyle = `rgb(${Math.round(r * 255)},${Math.round(g * 255)},${Math.round(b * 255)})`;
        ctx.fillRect(x, 0, 1, h);
    }
}

const palettePopup = document.createElement("div");
palettePopup.id = "palette-popup";

const popupParams = document.createElement("div");
popupParams.id = "palette-popup-params";
["a", "b", "c", "d"].forEach((key) => {
    const col = document.createElement("div");
    col.className = "palette-param-col";
    const input = document.createElement("input");
    input.type = "color";
    input.className = "palette-param-swatch";
    input.dataset.param = key;
    input.addEventListener("input", (e) => {
        const hex = e.target.value;
        const r = parseInt(hex.slice(1, 3), 16) / 255;
        const g = parseInt(hex.slice(3, 5), 16) / 255;
        const b = parseInt(hex.slice(5, 7), 16) / 255;
        currentPalette = { ...currentPalette, [key]: [r, g, b] };
        if (selectedPresetEl) {
            selectedPresetEl.classList.remove("selected");
            selectedPresetEl = null;
        }
        applyPalette(currentPalette);
    });
    const lbl = document.createElement("div");
    lbl.className = "palette-param-label";
    lbl.textContent = key;
    col.appendChild(input);
    col.appendChild(lbl);
    popupParams.appendChild(col);
});
palettePopup.appendChild(popupParams);

function updatePopupHeader() {
    const p = currentPalette;
    ["a", "b", "c", "d"].forEach((key) => {
        const input = palettePopup.querySelector(`.palette-param-swatch[data-param="${key}"]`);
        if (input) input.value = rgbToHex(p[key]);
    });
}

const popupPresets = document.createElement("div");
popupPresets.id = "palette-popup-presets";
let selectedPresetEl = null;

PALETTE_PRESETS.forEach((preset, idx) => {
    const row = document.createElement("div");
    row.className = "palette-preset" + (idx === 0 ? " selected" : "");
    if (idx === 0) selectedPresetEl = row;
    const canvas = document.createElement("canvas");
    canvas.width = 200;
    canvas.height = 16;
    row.appendChild(canvas);
    popupPresets.appendChild(row);
    requestAnimationFrame(() => drawPresetCanvas(canvas, preset));

    row.addEventListener("click", () => {
        if (selectedPresetEl) selectedPresetEl.classList.remove("selected");
        row.classList.add("selected");
        selectedPresetEl = row;
        applyPalette(preset);
    });
});
palettePopup.appendChild(popupPresets);
document.getElementById("canvas-container").appendChild(palettePopup);

updatePopupHeader();

paletteStrip.style.cursor = "pointer";
paletteStrip.addEventListener("click", (e) => {
    e.stopPropagation();
    const rect = paletteStrip.getBoundingClientRect();
    palettePopup.style.right = `${window.innerWidth - rect.right}px`;
    palettePopup.style.top = `${rect.bottom + 6}px`;
    palettePopup.classList.toggle("visible");
});

document.addEventListener("click", (e) => {
    if (!palettePopup.contains(e.target)) palettePopup.classList.remove("visible");
});

eventManager.addListener("shapes/detected", (event) => {
    const count = event.count;
    const names = event.names || [];
    paletteShapeCount = count;
    colorPickersContainer.innerHTML = "";

    if (count === 0) {
        colorPickerPanel.classList.remove("visible");
        colorPickerPanel.classList.remove("open");
        return;
    }

    for (let i = 1; i <= count; i++) {
        const row = document.createElement("div");
        row.className = "color-picker-row";
        row.dataset.shapeIndex = i;

        const label = document.createElement("label");
        label.textContent = names[i - 1] || `Shape ${i}`;

        const input = document.createElement("input");
        input.type = "color";
        input.value = rgbToHex(cosineRGB(shapeT(i), currentPalette));
        input.addEventListener("input", (e) => {
            const hex = e.target.value;
            const r = parseInt(hex.slice(1, 3), 16) / 255;
            const g = parseInt(hex.slice(3, 5), 16) / 255;
            const b = parseInt(hex.slice(5, 7), 16) / 255;
            setShapeColor(i, r, g, b);
        });

        const vis = document.createElement("input");
        vis.type = "checkbox";
        vis.checked = true;
        vis.addEventListener("change", (e) => setShapeVisible(i, e.target.checked));

        row.appendChild(label);
        row.appendChild(input);
        row.appendChild(vis);
        colorPickersContainer.appendChild(row);
    }

    for (let i = 1; i <= count; i++) {
        const [r, g, b] = cosineRGB(shapeT(i), currentPalette);
        setShapeColor(i, r, g, b);
    }

    colorPickerPanel.classList.add("visible");
});

// ── Animation & Audio Panel ──

const animationPanel = document.getElementById("animation-panel");
const animTimeInput = document.getElementById("animation-time-input");
const btnPauseAnim = document.getElementById("btn-pause-anim");
const btnRewindAnim = document.getElementById("btn-rewind-anim");
let animPaused = true;
let animTimeInterval = null;

document.getElementById("animation-panel-header").addEventListener("click", (e) => {
    if (e.target.classList.contains("panel-close")) {
        animationPanel.classList.remove("open");
        clearInterval(animTimeInterval);
        animTimeInterval = null;
    } else if (!animationPanel.classList.contains("open")) {
        animationPanel.classList.add("open");
        animTimeInterval = setInterval(() => {
            if (!animPaused) animTimeInput.value = getAnimTime().toFixed(2);
        }, 50);
    }
});

btnPauseAnim.addEventListener("click", () => {
    animPaused = !animPaused;
    setAnimPaused(animPaused);
    btnPauseAnim.innerHTML = animPaused
        ? `<span class="anim-glyph">\u25B6</span>`
        : `<span class="anim-glyph anim-glyph-low">\u23F8</span>`;
    animTimeInput.readOnly = animPaused;
    if (animPaused) {
        animTimeInput.value = getAnimTime().toFixed(2);
    } else {
        requestAnimationFrame(sliderAnimTick);
    }
});

btnRewindAnim.addEventListener("click", () => {
    setAnimTime(0);
    animTimeInput.value = "0.00";
});

animTimeInput.addEventListener("change", () => {
    const t = parseFloat(animTimeInput.value);
    if (!isNaN(t) && t >= 0) setAnimTime(t);
});

document.getElementById("audio-volume").addEventListener("input", (e) => {
    const pct = parseInt(e.target.value, 10);
    const volume = pct / 100;
    audioSystem.setEnabled(pct > 0);
    audioSystem.setVolume(volume);
});

eventManager.addListener("canvas/update-audio", (event) => {
    const { javascript, hasAudio } = event;
    const audioVolumeRow = document.getElementById("audio-volume-row");
    audioVolumeRow.style.display = hasAudio ? "" : "none";
    audioSystem.updateShader(javascript, hasAudio);
});

// ── Cross Section Panel ──

const fieldPanel = document.getElementById("field-panel");
const fieldHeightSlider = document.getElementById("field-height");
const fieldHeightLabel = document.getElementById("field-height-label");
const fieldModelSlider = document.getElementById("field-model");
const fieldModelLabel = document.getElementById("field-model-label");

document.getElementById("field-panel-header").addEventListener("click", (e) => {
    if (e.target.classList.contains("panel-close")) {
        fieldPanel.classList.remove("open");
        setFieldEnabled(false);
        setModelOpacity(1.0);
    } else if (!fieldPanel.classList.contains("open")) {
        fieldPanel.classList.add("open");
        setFieldEnabled(true);
        setModelOpacity(parseFloat(fieldModelSlider.value));
    }
});

fieldHeightSlider.addEventListener("input", (e) => {
    const value = parseFloat(e.target.value);
    fieldHeightLabel.textContent = value.toFixed(2);
    setFieldHeight(value);
});

const fieldMin = parseFloat(fieldHeightSlider.min);
const fieldMax = parseFloat(fieldHeightSlider.max);

function commitFieldHeight(text) {
    const value = parseFloat(text);
    if (isNaN(value)) return;
    const clamped = Math.min(Math.max(value, fieldMin), fieldMax);
    fieldHeightSlider.value = clamped;
    fieldHeightLabel.textContent = clamped.toFixed(2);
    setFieldHeight(clamped);
}

fieldHeightLabel.addEventListener("mousedown", () => { fieldHeightLabel.contentEditable = "true"; });
fieldHeightLabel.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        e.preventDefault();
        commitFieldHeight(fieldHeightLabel.textContent);
        fieldHeightLabel.contentEditable = "false";
        fieldHeightLabel.blur();
    } else if (e.key === "Escape") {
        fieldHeightLabel.textContent = parseFloat(fieldHeightSlider.value).toFixed(2);
        fieldHeightLabel.contentEditable = "false";
        fieldHeightLabel.blur();
    }
});
fieldHeightLabel.addEventListener("blur", () => {
    if (fieldHeightLabel.contentEditable === "true") {
        commitFieldHeight(fieldHeightLabel.textContent);
        fieldHeightLabel.contentEditable = "false";
    }
});

const fieldIntervalSlider = document.getElementById("field-interval");
const fieldIntervalLabel = document.getElementById("field-interval-label");

fieldIntervalSlider.addEventListener("input", (e) => {
    const value = parseFloat(e.target.value);
    fieldIntervalLabel.textContent = value.toFixed(2);
    if (value > 0) setFieldInterval(value);
});

const fieldIntervalMin = parseFloat(fieldIntervalSlider.min);
const fieldIntervalMax = parseFloat(fieldIntervalSlider.max);

function commitFieldInterval(text) {
    const value = parseFloat(text);
    if (isNaN(value)) return;
    const clamped = Math.min(Math.max(value, fieldIntervalMin), fieldIntervalMax);
    fieldIntervalSlider.value = clamped;
    fieldIntervalLabel.textContent = clamped.toFixed(2);
    if (clamped > 0) setFieldInterval(clamped);
}

fieldIntervalLabel.addEventListener("mousedown", () => { fieldIntervalLabel.contentEditable = "true"; });
fieldIntervalLabel.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        e.preventDefault();
        commitFieldInterval(fieldIntervalLabel.textContent);
        fieldIntervalLabel.contentEditable = "false";
        fieldIntervalLabel.blur();
    } else if (e.key === "Escape") {
        fieldIntervalLabel.textContent = parseFloat(fieldIntervalSlider.value).toFixed(2);
        fieldIntervalLabel.contentEditable = "false";
        fieldIntervalLabel.blur();
    }
});
fieldIntervalLabel.addEventListener("blur", () => {
    if (fieldIntervalLabel.contentEditable === "true") {
        commitFieldInterval(fieldIntervalLabel.textContent);
        fieldIntervalLabel.contentEditable = "false";
    }
});

const fieldFadeSlider = document.getElementById("field-fade");
const fieldFadeLabel = document.getElementById("field-fade-label");
const fieldFadeMin = parseFloat(fieldFadeSlider.min);
const fieldFadeMax = parseFloat(fieldFadeSlider.max);

function fadeSliderToUniform(v) { return v >= fieldFadeMax ? 1e9 : v; }

fieldFadeSlider.addEventListener("input", (e) => {
    const value = parseFloat(e.target.value);
    fieldFadeLabel.textContent = value >= fieldFadeMax ? "\u221E" : value.toFixed(1);
    setFieldFade(fadeSliderToUniform(value));
});

function commitFieldFade(text) {
    const value = parseFloat(text);
    if (isNaN(value)) return;
    const clamped = Math.min(Math.max(value, fieldFadeMin), fieldFadeMax);
    fieldFadeSlider.value = clamped;
    fieldFadeLabel.textContent = clamped >= fieldFadeMax ? "\u221E" : clamped.toFixed(1);
    setFieldFade(fadeSliderToUniform(clamped));
}

fieldFadeLabel.addEventListener("mousedown", () => { fieldFadeLabel.contentEditable = "true"; });
fieldFadeLabel.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        e.preventDefault();
        commitFieldFade(fieldFadeLabel.textContent);
        fieldFadeLabel.contentEditable = "false";
        fieldFadeLabel.blur();
    } else if (e.key === "Escape") {
        const v = parseFloat(fieldFadeSlider.value);
        fieldFadeLabel.textContent = v >= fieldFadeMax ? "\u221E" : v.toFixed(1);
        fieldFadeLabel.contentEditable = "false";
        fieldFadeLabel.blur();
    }
});
fieldFadeLabel.addEventListener("blur", () => {
    if (fieldFadeLabel.contentEditable === "true") {
        commitFieldFade(fieldFadeLabel.textContent);
        fieldFadeLabel.contentEditable = "false";
    }
});

fieldModelSlider.addEventListener("input", (e) => {
    const value = parseFloat(e.target.value);
    fieldModelLabel.textContent = value.toFixed(2);
    setModelOpacity(value);
});

function commitFieldModel(text) {
    const value = parseFloat(text);
    if (isNaN(value)) return;
    const clamped = Math.min(Math.max(value, 0), 1);
    fieldModelSlider.value = clamped;
    fieldModelLabel.textContent = clamped.toFixed(2);
    setModelOpacity(clamped);
}

fieldModelLabel.addEventListener("mousedown", () => { fieldModelLabel.contentEditable = "true"; });
fieldModelLabel.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        e.preventDefault();
        commitFieldModel(fieldModelLabel.textContent);
        fieldModelLabel.contentEditable = "false";
        fieldModelLabel.blur();
    } else if (e.key === "Escape") {
        fieldModelLabel.textContent = parseFloat(fieldModelSlider.value).toFixed(2);
        fieldModelLabel.contentEditable = "false";
        fieldModelLabel.blur();
    }
});
fieldModelLabel.addEventListener("blur", () => {
    if (fieldModelLabel.contentEditable === "true") {
        commitFieldModel(fieldModelLabel.textContent);
        fieldModelLabel.contentEditable = "false";
    }
});

const fieldSectionSlider = document.getElementById("field-section");
const fieldSectionLabel = document.getElementById("field-section-label");

fieldSectionSlider.addEventListener("input", (e) => {
    const value = parseFloat(e.target.value);
    fieldSectionLabel.textContent = value.toFixed(2);
    setSectionOpacity(value);
});

function commitFieldSection(text) {
    const value = parseFloat(text);
    if (isNaN(value)) return;
    const clamped = Math.min(Math.max(value, 0), 1);
    fieldSectionSlider.value = clamped;
    fieldSectionLabel.textContent = clamped.toFixed(2);
    setSectionOpacity(clamped);
}

fieldSectionLabel.addEventListener("mousedown", () => { fieldSectionLabel.contentEditable = "true"; });
fieldSectionLabel.addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        e.preventDefault();
        commitFieldSection(fieldSectionLabel.textContent);
        fieldSectionLabel.contentEditable = "false";
        fieldSectionLabel.blur();
    } else if (e.key === "Escape") {
        fieldSectionLabel.textContent = parseFloat(fieldSectionSlider.value).toFixed(2);
        fieldSectionLabel.contentEditable = "false";
        fieldSectionLabel.blur();
    }
});
fieldSectionLabel.addEventListener("blur", () => {
    if (fieldSectionLabel.contentEditable === "true") {
        commitFieldSection(fieldSectionLabel.textContent);
        fieldSectionLabel.contentEditable = "false";
    }
});

document.getElementById("field-union").addEventListener("change", (e) => {
    setFieldUnion(e.target.checked);
});

document.getElementById("field-external").addEventListener("change", (e) => {
    setFieldExternal(e.target.checked);
});

function updateFieldNormal() {
    const x = parseFloat(document.getElementById("field-normal-x").value) || 0;
    const y = parseFloat(document.getElementById("field-normal-y").value) || 0;
    const z = parseFloat(document.getElementById("field-normal-z").value) || 0;
    setFieldNormal(x, y, z);
}
document.getElementById("field-normal-x").addEventListener("input", updateFieldNormal);
document.getElementById("field-normal-y").addEventListener("input", updateFieldNormal);
document.getElementById("field-normal-z").addEventListener("input", updateFieldNormal);

// ── Top Bar: Example Dropdown & Initial Load ──

async function loadExample(name) {
    if (!name) return;
    try {
        const res = await fetch(`/api/example?name=${encodeURIComponent(name)}`);
        const data = await res.json();
        if (!res.ok || data.error) {
            throw new Error(data.error || `HTTP ${res.status}`);
        }
        currentExampleName = data.name || name;
        exampleSelectEl.value = currentExampleName;
        selectTab("GSL");
        const code = (data.source || "").trim();
        suppressDirty = true;
        setContent(code);
        suppressDirty = false;
        gslSource = code;
        currentGSL = code;
        setDirty(false);
        await compileGsl(code);
    } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        showNotification(true, `Failed to load ${name}: ${msg}`);
    }
}

async function refreshExampleList(selectedName = "") {
    try {
        const res = await fetch("/api/examples");
        const data = await res.json();
        const examples = (data && Array.isArray(data.examples)) ? data.examples : [];
        exampleSelectEl.innerHTML = "";
        for (const ex of examples) {
            const opt = document.createElement("option");
            opt.value = ex.name;
            opt.textContent = ex.name;
            exampleSelectEl.appendChild(opt);
        }
        if (selectedName && examples.some((e) => e.name === selectedName)) {
            exampleSelectEl.value = selectedName;
        }
        return examples;
    } catch (err) {
        console.error("Failed to fetch /api/examples:", err);
        return [];
    }
}

exampleSelectEl.addEventListener("change", () => {
    const selected = exampleSelectEl.value;
    if (selected) {
        loadExample(selected);
    }
});

(async function initExamples() {
    const examples = await refreshExampleList();
    if (examples.length > 0) {
        const defaultEx = examples.find((e) => e.name === "gear.gsl") || examples[0];
        await loadExample(defaultEx.name);
    }
})();
