(function () {

    const ANSI = {
        reset: '\x1b[0m',

        bold: '\x1b[1m',

        green: '\x1b[32m',
        red: '\x1b[31m',
        yellow: '\x1b[33m',
        cyan: '\x1b[36m',
        gray: '\x1b[38;5;250m'
    }

    function color(text, ...styles) {
        return styles.join('') + text + ANSI.reset
    }

    function createTerminal(el) {
        const term = new Terminal({
            cols: 80,
            rows: 20,
            convertEol: true,
            cursorBlink: true,
            fontFamily: "'Roboto Mono', 'Nerd Symbols', monospace",
            fontSize: 14
        })

        el.innerHTML = ''
        term.open(el)

        return term
    }

    function runWasm(el, wasmFile) {
        const term = createTerminal(el)

        term.writeln(color(`▶ Starting ${wasmFile}...\n`, ANSI.gray))

        const worker = new Worker('_wasm/wasm-worker.js')

        worker.onmessage = e => {
            const msg = e.data

            if (msg.type === 'stdout') {
                term.write(msg.data)
            }

            if (msg.type === 'error') {
                term.writeln('\n' + color('✖ error\n' + msg.data, ANSI.bold, ANSI.red))
            }

            if (msg.type === 'done') {
                term.writeln('\n' + color('✓ Finished', ANSI.gray))
            }
        }

        worker.postMessage({ wasm: wasmFile })
    }

    function autoRun() {
        const nodes = document.querySelectorAll('.wasm-terminal')

        nodes.forEach(el => {
            const wasm = el.dataset.wasm

            if (!wasm) return

            // prevent double execution
            if (el.__wasmStarted) return
            el.__wasmStarted = true

            runWasm(el, wasm)
        })
    }

    window.$docsify = window.$docsify || {}
    window.$docsify.plugins = [].concat(window.$docsify.plugins || [], function (hook) {

        hook.doneEach(function () {
            // wait for DOM render
            setTimeout(autoRun, 100)
        })

        // optional: reset flags on navigation
        hook.beforeEach(function () {
            document.querySelectorAll('.wasm-terminal').forEach(el => {
                el.__wasmStarted = false
            })
        })

    })
})()