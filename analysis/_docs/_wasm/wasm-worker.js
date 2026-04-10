importScripts('wasm_exec.js')

self.onmessage = async (e) => {
    const { wasm } = e.data

    const go = new Go()
    const decoder = new TextDecoder('utf-8')

    if (!self.fs) {
        self.fs = {}
    }

    const originalWriteSync = self.fs.writeSync || function (fd, buf) {
        return buf.length
    };

    self.fs.writeSync = function (fd, buf) {
        if (fd === 1) {
            self.postMessage({
                type: 'stdout',
                data: decoder.decode(buf),
            })
        }
        return originalWriteSync(fd, buf)
    }

    try {
        const result = await WebAssembly.instantiateStreaming(fetch(wasm), go.importObject)
        await go.run(result.instance)
        self.postMessage({ type: 'done' })
    } catch (err) {
        self.postMessage({
            type: 'error',
            data: err.toString(),
        })
    }
}