// A wasm plugin for docsify that HTML (<div>) elements with class
// .wasm-terminal into live terminals rendering the terminal output of wasm
// programs executing in background workers.
//
// See also:
// https://github.com/docsifyjs/docsify/blob/develop/docs/write-a-plugin.md
(function () {

    const ANSI = {
        reset: '\x1b[0m',
        bold: '\x1b[1m',
        green: '\x1b[32m',
        red: '\x1b[31m',
        yellow: '\x1b[33m',
        cyan: '\x1b[36m',
        gray: '\x1b[38;5;250m',
    }

    function colorize(text, ...styles) {
        return styles.join('') + text + ANSI.reset
    }

    function parseBool(val, def) {
        if (val === undefined) return !!def
        return (val === 'true') || (val === 'on') || (val === '1')
    }

    function getConfig(el) {
        return {
            rows: parseInt(el.dataset.rows || '20', 10),
            cursor: el.dataset.cursor || 'block',
            blink: parseBool(el.dataset.blink),
        }
    }

    function createTerminal(el, config) {
        const term = new Terminal({
            cols: 80,
            rows: config.rows,
            scrollback: 500,
            convertEol: true,
            cursorStyle: config.cursor,
            cursorBlink: config.blink,
            fontFamily: "'Roboto Mono', 'Nerd Symbols Mono', monospace",
            fontSize: 14,
            lineHeight: 1.1,
            letterSpacing: 0,
        })
        el.innerHTML = ''
        term.open(el)
        return term
    }

    function runWasm(el, wasmFile) {
        const config = getConfig(el)
        el.innerHTML = '' // clear the terminal container

        // we next want to wrap the terminal
        const container = document.createElement('div')
        const toolbar = document.createElement('div')
        const termEl = document.createElement('div')

        toolbar.style.display = 'flex'
        toolbar.style.justifyContent = 'flex-end'
        toolbar.style.marginBottom = '4px'

        // Run button
        const btn = document.createElement('button')
        btn.className = 'wasm-runagain'
        btn.textContent = '󰑓 Run again'

        toolbar.appendChild(btn)
        container.appendChild(toolbar)
        container.appendChild(termEl)
        el.appendChild(container)

        let worker = null
        function run() {
            if (worker) {
                worker.terminate()
                worker = null
            }
            termEl.innerHTML = ''
            const term = createTerminal(termEl, config)

            term.writeln(colorize(` Running ${wasmFile}...`, ANSI.gray))

            worker = new Worker('/_wasm/wasm-worker.js')
            worker.onmessage = e => {
                const msg = e.data
                switch (msg.type) {
                    case 'stdout':
                        term.write(msg.data)
                        break
                    case 'error':
                        if (term.buffer.active.cursorX !== 0) {
                            term.write('\n')
                        }
                        term.writeln(colorize('✖ error\n' + msg.data, ANSI.bold, ANSI.red))
                        break
                    case 'done':
                        if (term.buffer.active.cursorX !== 0) {
                            term.write('\n')
                        }
                        term.write(colorize(' Finished', ANSI.gray))
                        worker.terminate()
                        worker = null
                        break
                }
            }
            worker.postMessage({ wasm: wasmFile })
        }

        btn.onclick = run
        run()

        el.__terminateworker = () => {
            if (worker) worker.terminate()
        }
    }

    function autoRun() {
        document.querySelectorAll('.wasm-terminal').forEach(el => {
            const wasm = el.dataset.wasm
            if (!wasm || el.__wasmStarted) return
            el.__wasmStarted = true
            runWasm(el, wasm)
        })
    }

    function wasmPlugin(hook, vm) {
        // Invoked on each page load before new markdown is transformed to HTML.
        // Supports asynchronous tasks (see beforeEach documentation for details).
        hook.beforeEach(function () {
            document.querySelectorAll('.wasm-terminal').forEach(el => {
                el.__wasmStarted = false
                if (el.__terminateworker) el.__terminateworker()
            })
        })
        // Invoked on each page load after new HTML has been appended to the DOM
        hook.doneEach(function () {
            autoRun() // setTimeout(autoRun, 100)
        })
    }

    // Finally add our plugin to docsify's plugin array...
    window.$docsify = window.$docsify || {}
    window.$docsify.plugins = [].concat(window.$docsify.plugins || [], wasmPlugin)
})()