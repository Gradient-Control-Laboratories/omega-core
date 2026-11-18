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
