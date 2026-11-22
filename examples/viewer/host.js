// The viewer's host: where documents come from, what compiles them, and where they are saved.
//
// This is the one `gsl serve` and `gsl viewer` use: each call goes to the local server's /api.
// A page written by `gsl export` has host-static.js in this file's place, which answers the
// same calls from the files saved beside the page and compiles nothing. main.js reads
// `editable` to decide what the page offers.

// editable is true when the host compiles: the document can be edited, saved and meshed.
export const editable = true;

async function jsonOrThrow(res) {
    const data = await res.json();
    if (!res.ok || data.error) {
        throw new Error(data.error || `HTTP ${res.status}`);
    }
    return data;
}

// listExamples is the documents the host offers, as [{ name, label }].
export async function listExamples() {
    const data = await (await fetch("/api/examples")).json();
    return (data && Array.isArray(data.examples)) ? data.examples : [];
}

// loadExample is one document: { name, source }.
export async function loadExample(name) {
    return jsonOrThrow(await fetch(`/api/example?name=${encodeURIComponent(name)}`));
}

// saveExample writes source under name and answers { name }, the name it was saved as.
export async function saveExample(name, source) {
    return jsonOrThrow(await fetch("/api/save", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, source }),
    }));
}

// compile compiles request ({ id, clientId, source }) and answers { ok, status, data }, where
// data is the field bundle: the WGSL and what the viewer draws it with, or the document's errors.
export async function compile(request) {
    const res = await fetch("/api/compile", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(request),
    });
    return { ok: res.ok, status: res.status, data: await res.json() };
}

// mesh meshes request's source on the server and answers the HTTP response, whose body is the
// file and whose Content-Type names its format.
export function mesh(request) {
    return fetch("/api/mesh", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(request),
    });
}

// renderLit asks the server for the lit renderer's frame of request's document from the viewer's
// camera ({ id, clientId, source, width, height, worldExtent, time, camera, colors, mode, simple,
// backDim }) and answers { width, height, data }: the frame's size, and its pixels as five float32s
// each, row-major from the top left (the colour, the return hit, 0 for none, and the depth). It
// takes as long as the lit renderer does, and reports how far it has got on the progress stream
// (connectProgress) as stage "render".
export async function renderLit(request) {
    const res = await fetch("/api/render-lit", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(request),
    });
    if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || `HTTP ${res.status}`);
    }
    return {
        width: Number(res.headers.get("X-Frame-Width")),
        height: Number(res.headers.get("X-Frame-Height")),
        data: new Float32Array(await res.arrayBuffer()),
    };
}

// tape asks the server for request's document ({ id, clientId, source, appearance }) as an
// evaluation tape and answers its bytes, as a Uint8Array: every return of the document, and with
// appearance each return's colour and material fields. The Fidget device draws from it.
export async function tape(request) {
    const res = await fetch("/api/tape", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(request),
    });
    if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || `HTTP ${res.status}`);
    }
    return new Uint8Array(await res.arrayBuffer());
}

// exportPage asks for request's document ({ name, source }) as `gsl export` writes it, and answers
// the HTTP response, whose body is a zip of the folder: this viewer with the static host in this
// file's place, and the document's field beside it.
export function exportPage(request) {
    return fetch("/api/export", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(request),
    });
}

// uploadAsset sends an imported mesh's bytes to the server, so a program that reads the mesh
// compiles there.
export async function uploadAsset(name, bytes) {
    const res = await fetch(`/api/asset?name=${encodeURIComponent(name)}`, {
        method: "POST",
        headers: { "Content-Type": "application/octet-stream" },
        body: bytes,
    });
    if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || `upload failed (HTTP ${res.status})`);
    }
}

// connectProgress reads the server's NDJSON progress stream for clientId and hands each event
// to onEvent. It reconnects when the stream closes, and never returns.
export async function connectProgress(clientId, onEvent) {
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
                        onEvent(JSON.parse(line));
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
