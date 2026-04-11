(function () {

    const MAGIC_PREFIX = 'code-anchor:'

    function extractId(info) {
        if (!info) return ''
        
        const patterns = [
            /\bid="([^"]+)"/,   // id="..."
            /\bid='([^']+)'/,   // id='...'
            /\bid=([^\s]+)/     // id=...
        ]
        for (const re of patterns) {
            const match = info.match(re)
            if (match) return match[1]
        }
        return ''
    }

    function removeId(info) {
        if (!info) return ''

        return info
            .replace(/\bid="[^"]+"/g, '')
            .replace(/\bid='[^']+'/g, '')
            .replace(/\bid=[^\s]+/g, '')
            .trim()
    }

    function extractLang(info) {
        if (!info) return ''

        const cleaned = removeId(info)
        const parts = cleaned.split(/\s+/)
        return parts[0] || ''
    }

    function install(hook, vm) {

        hook.beforeEach(function (md) {
            return md.replace(/```([^\n]*)\n([\s\S]*?)```/g, function (full, info, code) {
                const id = extractId(info)
                if (!id) return full

                const lang = extractLang(info)
                const cleanFence = `\`\`\`${lang}\n${code}\`\`\``
                return `<!--${MAGIC_PREFIX}${id}-->
${cleanFence}`
            })
        })

        hook.doneEach(function () {
            const walker = document.createTreeWalker(
                document.body,
                NodeFilter.SHOW_COMMENT,
                null,
                false
            );

            let node
            while ((node = walker.nextNode())) {
                if (!node.nodeValue.startsWith(MAGIC_PREFIX)) continue

                const id = node.nodeValue.slice(MAGIC_PREFIX.length)
                let el = node.nextSibling
                while (el) {
                    if (el.nodeType === 1 && el.tagName === 'PRE') {
                        const code = el.querySelector('code')
                        if (code && !code.id) {
                            code.id = id
                        }
                        break
                    }
                    el = el.nextSibling
                }
                node.remove()
            }
        })
    }

    window.$docsify = window.$docsify || {}
    window.$docsify.plugins = [].concat(install, window.$docsify.plugins || [])
})()
