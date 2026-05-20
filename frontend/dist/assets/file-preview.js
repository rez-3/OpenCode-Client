// ============================================================
// 站内文件浏览器 - 预览器分发
// ============================================================

function fileBrowserBuildApiURL(kind, rootDir, relPath) {
    return '/api/files/' + kind + '?rootDir=' + encodeURIComponent(rootDir || '') + '&path=' + encodeURIComponent(relPath || '/');
}

function fileBrowserBuildRawURL(rootDir, relPath) {
    return fileBrowserBuildApiURL('raw', rootDir, relPath);
}

function fileBrowserIsMarkdown(ext) {
    return ext === '.md' || ext === '.markdown';
}

function fileBrowserIsImage(ext) {
    return ['.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg'].indexOf(ext) >= 0;
}

function fileBrowserIsPDF(ext) {
    return ext === '.pdf';
}

function fileBrowserIsCSV(ext) {
    return ext === '.csv';
}

function fileBrowserIsSpreadsheet(ext) {
    return ext === '.xlsx' || ext === '.xls';
}

function fileBrowserIsText(ext) {
    return ['.txt', '.log', '.json', '.yaml', '.yml', '.ini', '.env', '.xml', '.js', '.ts', '.tsx', '.jsx', '.go','.sum','.mod','.py', '.java', '.c', '.cpp', '.cc', '.rs', '.sh', '.bash', '.css', '.scss', '.less', '.html', '.htm', '.sql','.bat','.sh'].indexOf(ext) >= 0;
}

function fileBrowserEscapeHTML(text) {
    if (text == null) return '';
    var div = document.createElement('div');
    div.textContent = String(text);
    return div.innerHTML;
}

function fileBrowserSanitizeMarkedHtml(html) {
    var template = document.createElement('template');
    template.innerHTML = html;
    var allowedTags = new Set(['A', 'P', 'BR', 'STRONG', 'EM', 'CODE', 'PRE', 'UL', 'OL', 'LI', 'BLOCKQUOTE', 'H1', 'H2', 'H3', 'H4', 'H5', 'H6', 'HR', 'TABLE', 'THEAD', 'TBODY', 'TR', 'TH', 'TD']);
    sanitizeNodeTree(template.content, allowedTags);
    return template.innerHTML;
}

function sanitizeNodeTree(root, allowedTags) {
    var children = Array.prototype.slice.call(root.childNodes || []);
    children.forEach(function(node) {
        if (node.nodeType === Node.TEXT_NODE) return;
        if (node.nodeType !== Node.ELEMENT_NODE) {
            root.removeChild(node);
            return;
        }
        if (!allowedTags.has(node.tagName)) {
            var text = document.createTextNode(node.textContent || '');
            root.replaceChild(text, node);
            return;
        }
        var attrs = Array.prototype.slice.call(node.attributes || []);
        attrs.forEach(function(attr) {
            var attrName = attr.name.toLowerCase();
            if (node.tagName === 'A' && attrName === 'href') {
                var href = (attr.value || '').trim();
                if (/^(https?:|mailto:|#|\/)/i.test(href)) {
                    node.setAttribute('target', '_blank');
                    node.setAttribute('rel', 'noopener noreferrer');
                } else {
                    node.removeAttribute(attr.name);
                }
                return;
            }
            node.removeAttribute(attr.name);
        });
        sanitizeNodeTree(node, allowedTags);
    });
}

function fileBrowserFormatBytes(bytes) {
    if (!bytes) return '0 B';
    var units = ['B', 'KB', 'MB', 'GB'];
    var i = 0;
    var val = bytes;
    while (val >= 1024 && i < units.length - 1) {
        val /= 1024;
        i++;
    }
    return (i === 0 ? val : val.toFixed(1)) + ' ' + units[i];
}

async function renderFilePreview(item) {
    var state = window.fileBrowserState;
    if (!state || !item || item.type !== 'file') return;
    state.selectedItem = item;
    var titleEl = document.getElementById('filePreviewTitle');
    var metaEl = document.getElementById('filePreviewMeta');
    var bodyEl = document.getElementById('filePreviewBody');
    if (titleEl) titleEl.textContent = item.name;
    if (metaEl) metaEl.textContent = '加载中...';
    if (bodyEl) bodyEl.innerHTML = '<div class="file-browser-empty">正在读取文件...</div>';

    try {
        var statResp = await fetch(fileBrowserBuildApiURL('stat', state.rootDir, item.path));
        if (!statResp.ok) throw new Error('读取文件信息失败');
        var meta = await statResp.json();
        state.previewMeta = meta;
        if (metaEl) {
            metaEl.textContent = [meta.ext || '', fileBrowserFormatBytes(meta.size || 0), meta.modifiedAt || ''].filter(Boolean).join(' · ');
        }
        var ext = (meta.ext || '').toLowerCase();

        if (fileBrowserIsImage(ext)) {
            bodyEl.innerHTML = '<div class="file-browser-image-wrap"><img class="file-browser-image" src="' + fileBrowserBuildRawURL(state.rootDir, item.path) + '" alt="' + fileBrowserEscapeHTML(item.name) + '"></div>';
            return;
        }
        if (fileBrowserIsPDF(ext)) {
            bodyEl.innerHTML = '<iframe class="file-browser-pdf" title="PDF预览" src="' + fileBrowserBuildRawURL(state.rootDir, item.path) + '"></iframe>';
            return;
        }

        if (fileBrowserIsSpreadsheet(ext)) {
            bodyEl.innerHTML = '<div class="file-browser-unsupported">' +
                '<p>当前版本未启用 Excel 在线预览。</p>' +
                '<p>文件：' + fileBrowserEscapeHTML(item.name) + '</p>' +
                '<a class="btn btn-sm btn-add" href="' + fileBrowserBuildRawURL(state.rootDir, item.path) + '" target="_blank" rel="noopener noreferrer">下载文件</a>' +
                '</div>';
            return;
        }

        if (fileBrowserIsCSV(ext) || fileBrowserIsMarkdown(ext) || fileBrowserIsText(ext)) {
            var readResp = await fetch(fileBrowserBuildApiURL('read', state.rootDir, item.path));
            if (!readResp.ok) throw new Error('读取文件内容失败');
            var readData = await readResp.json();
            state.previewContent = readData.content || '';

            if (fileBrowserIsMarkdown(ext)) {
                bodyEl.innerHTML = '<div class="oc-text file-browser-markdown">' + fileBrowserSanitizeMarkedHtml(marked.parse(readData.content || '')) + '</div>';
                return;
            }
            if (fileBrowserIsCSV(ext)) {
                bodyEl.innerHTML = renderCSVPreview(readData.content || '');
                return;
            }
            bodyEl.innerHTML = '<pre class="file-browser-code"><code>' + fileBrowserEscapeHTML(readData.content || '') + '</code></pre>';
            return;
        }

        bodyEl.innerHTML = '<div class="file-browser-unsupported">' +
            '<p>该文件类型暂不支持在线预览。</p>' +
            '<p>文件：' + fileBrowserEscapeHTML(item.name) + '</p>' +
            '<a class="btn btn-sm btn-add" href="' + fileBrowserBuildRawURL(state.rootDir, item.path) + '" target="_blank" rel="noopener noreferrer">下载文件</a>' +
            '</div>';
    } catch (err) {
        if (metaEl) metaEl.textContent = '';
        if (bodyEl) bodyEl.innerHTML = '<div class="file-browser-empty error">' + fileBrowserEscapeHTML(err.message || err) + '</div>';
    }
}

function renderCSVPreview(content) {
    var lines = String(content || '').split(/\r?\n/).filter(function(line) { return line !== ''; });
    if (!lines.length) return '<div class="file-browser-empty">CSV 文件为空</div>';
    var rows = lines.map(function(line) { return line.split(','); });
    var html = '<div class="file-browser-table-wrap"><table class="file-browser-table"><tbody>';
    rows.forEach(function(cols, rowIndex) {
        html += '<tr>';
        cols.forEach(function(col) {
            if (rowIndex === 0) {
                html += '<th>' + fileBrowserEscapeHTML(col) + '</th>';
            } else {
                html += '<td>' + fileBrowserEscapeHTML(col) + '</td>';
            }
        });
        html += '</tr>';
    });
    html += '</tbody></table></div>';
    return html;
}
