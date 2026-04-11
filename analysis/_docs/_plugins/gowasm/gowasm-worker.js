importScripts('wasm_exec.js')

self.onmessage = async (e) => {

    function simpleParseArgs(str) {
        if (!str) return []

        const re = /"([^"]*)"|'([^']*)'|[^\s]+/g
        const args = []
        let match

        while ((match = re.exec(str)) !== null) {
            args.push(match[1] ?? match[2] ?? match[0])
        }

        return args
    }

    const { wasm, args, codearg } = e.data

    const go = new Go()
    go.argv = [
        wasm, 
        ...simpleParseArgs(args),
        ...(codearg ? [codearg] : []),
    ]

    const decoder = new TextDecoder('utf-8')

    if (!self.fs) {
        self.fs = {}
    }

    const originalWriteSync = self.fs.writeSync || function (fd, buf) {
        return buf.length
    }

    self.fs.writeSync = function (fd, buf) {
        if (fd === 1) {
            self.postMessage({
                type: 'stdout',
                data: decoder.decode(buf),
            })
            return buf.length // don't chain
        } else if (fd === 2) {
            self.postMessage({
                type: 'stderr',
                data: decoder.decode(buf),
            })
            return buf.length // don't chain
        }
        return originalWriteSync(fd, buf)
    }

    // Try to fetch and stream the specified wasm file; if the server correctly
    // handles "Content-Encoding: gzip" with "Content-Type: application/wasm"
    // then we should be fine in this first attempt, unless that hits the wall.
    try {
        const result = await WebAssembly.instantiateStreaming(fetch(wasm), go.importObject)
        await go.run(result.instance)
        self.postMessage({ type: 'done' })
    } catch (err) {
        // Try again, this time with explicit .wasm.gz fetch and see if we're
        // going to be lucky. The downside is that we need to decompress
        // explicitly before we can instantiate, but that's a price to pay for
        // transparent handling of especially the bare-bones docsify-cli server.
        try {
            const res = await fetch(wasm + ".gz");
            const compressed = await res.arrayBuffer();

            const ds = new DecompressionStream("gzip");
            const stream = new Blob([compressed]).stream().pipeThrough(ds);
            const buffer = await new Response(stream).arrayBuffer();

            const result = await WebAssembly.instantiate(buffer, go.importObject);
            await go.run(result.instance)
            self.postMessage({ type: 'done' })
        } catch (gzerr) {
            self.postMessage({
                type: 'error',
                data: 'first error:\n' + err.toString() + '\nsecond error:\n' + gzerr.toString(),
            })
        }
    }
}