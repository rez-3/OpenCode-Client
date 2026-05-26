// ============================================================
// 站内文件浏览器 - 预览器分发
// ============================================================

async function fileBrowserApiStat(rootDir, relPath) {
    return await api.StatBrowserFile(rootDir, relPath);
}

async function fileBrowserApiRead(rootDir, relPath) {
    return await api.ReadBrowserFile(rootDir, relPath);
}

async function fileBrowserApiSave(rootDir, relPath, content) {
    return await api.SaveBrowserFile(rootDir, relPath, content);
}

async function fileBrowserApiGitPreview(rootDir, relPath) {
    return await api.GetGitPreview(rootDir, relPath);
}

async function fileBrowserApiGitHistoryPreview(rootDir, commitHash, relPath) {
    return await api.GetGitHistoryPreview(rootDir, commitHash, relPath);
}

function fileBrowserClearObjectURL() {
    var state = window.fileBrowserState;
    if (state && state.previewObjectURL) {
        URL.revokeObjectURL(state.previewObjectURL);
        state.previewObjectURL = '';
    }
}

function base64ToBlob(base64, mime) {
    var binary = atob(base64 || '');
    var len = binary.length;
    var bytes = new Uint8Array(len);
    for (var i = 0; i < len; i++) bytes[i] = binary.charCodeAt(i);
    return new Blob([bytes], { type: mime || 'application/octet-stream' });
}

async function fileBrowserResolveRawResource(rootDir, relPath) {
    var raw = await api.ReadBrowserRawBase64(rootDir, relPath);
    var blob = base64ToBlob(raw.base64 || '', raw.mime || 'application/octet-stream');
    var url = URL.createObjectURL(blob);
    window.fileBrowserState.previewObjectURL = url;
    return { url: url, name: raw.name || '', mime: raw.mime || 'application/octet-stream' };
}

function fileBrowserEscapeHTML(text) {
    if (text == null) return '';
    var div = document.createElement('div');
    div.textContent = String(text);
    return div.innerHTML;
}

function fileBrowserHighlightCode(code, ext) {
    var lines = String(code || '').split('\n');
    var numbered = '';
    for (var i = 0; i < lines.length; i++) {
        numbered += '<div class="hljs-line"><span class="hljs-line-no">' + (i + 1) + '</span><span class="hljs-line-content">' + (fileBrowserEscapeHTML(lines[i]) || ' ') + '</span></div>';
    }
    return numbered;
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

function updateFileBrowserDownloadButton(item) {
    var path = item && item.path ? item.path : '';
    var name = item && item.name ? item.name : '';
    if (typeof setFileBrowserDownloadTarget === 'function') {
        setFileBrowserDownloadTarget(path, name);
    }
}

function fileBrowserCanEdit(meta) {
    return !!(meta && meta.editable);
}

function isFileBrowserDarkTheme(theme) {
    var current = theme || document.documentElement.getAttribute('data-theme') || 'dark';
    return current === 'dark';
}

function destroyFileBrowserEditor() {
    var state = window.fileBrowserState;
    stopFileBrowserSearchButtonSync();
    if (state && state.previewEditorInstance && window.ProjectConfigCodeEditor) {
        window.ProjectConfigCodeEditor.destroy(state.previewEditorInstance);
    }
    if (state) {
        state.previewEditorInstance = null;
    }
}

function syncFileBrowserEditorTheme(theme) {
    var state = window.fileBrowserState;
    if (state && state.previewEditorInstance && window.ProjectConfigCodeEditor) {
        window.ProjectConfigCodeEditor.setTheme(state.previewEditorInstance, isFileBrowserDarkTheme(theme));
    }
}

window.syncFileBrowserEditorTheme = syncFileBrowserEditorTheme;

function fileBrowserIsSearchOpen() {
    var state = window.fileBrowserState;
    return !!(state && state.previewEditorInstance && window.ProjectConfigCodeEditor && window.ProjectConfigCodeEditor.isSearchOpen(state.previewEditorInstance));
}

function refreshFileBrowserSearchButtonState() {
    var searchBtn = document.getElementById('btnFilePreviewSearch');
    if (!searchBtn) return;
    var isOpen = fileBrowserIsSearchOpen();
    searchBtn.classList.toggle('active', isOpen);
    searchBtn.textContent = isOpen ? '关闭搜索' : '搜索';
}

function stopFileBrowserSearchButtonSync() {
    var state = window.fileBrowserState;
    if (state && state.previewSearchSyncTimer) {
        clearInterval(state.previewSearchSyncTimer);
        state.previewSearchSyncTimer = null;
    }
}

function startFileBrowserSearchButtonSync() {
    var state = window.fileBrowserState;
    stopFileBrowserSearchButtonSync();
    if (!state || !state.previewEditorInstance) return;
    state.previewSearchSyncTimer = setInterval(refreshFileBrowserSearchButtonState, 200);
}

function fileBrowserCanPreview(meta) {
    if (!meta) return false;
    if (meta.previewKind === 'binary') return false;
    if (meta.previewKind === 'markdown') return true;
    if (fileBrowserCanEdit(meta)) return false;
    return !!meta.previewable || !meta.ext;
}

function fileBrowserGetPreferredRenderMode(meta) {
    if (!meta) return 'preview';
    if (fileBrowserCanEdit(meta)) return 'edit';
    return 'preview';
}

function fileBrowserIsDirty() {
    var state = window.fileBrowserState;
    return (state.previewEditorValue || '') !== (state.previewOriginalContent || '');
}

function renderFilePreviewToolbar() {
    var state = window.fileBrowserState;
    var actionsEl = document.getElementById('filePreviewActions');
    if (!actionsEl) return;

    var meta = state.previewMeta;
    if (!meta || state.previewMode !== 'file' || !state.selectedItem || state.selectedItem.type !== 'file') {
        actionsEl.innerHTML = '';
        return;
    }

    var canEdit = fileBrowserCanEdit(meta);
    var canPreview = fileBrowserCanPreview(meta);
    var allowModeToggle = !!(meta && meta.previewKind === 'markdown');
    var buttons = '';

    if (allowModeToggle && canPreview) {
        buttons += '<button type="button" class="btn btn-sm' + (state.previewRenderMode === 'preview' ? ' active-file-preview-action' : '') + '" id="btnFilePreviewModePreview">预览</button>';
    }
    if (canEdit) {
        if (allowModeToggle) {
            buttons += '<button type="button" class="btn btn-sm' + (state.previewRenderMode === 'edit' ? ' active-file-preview-action' : '') + '" id="btnFilePreviewModeEdit">编辑</button>';
        }
        if (state.previewRenderMode === 'edit') {
            buttons += '<button type="button" class="btn btn-sm pc-editor-search-btn' + (fileBrowserIsSearchOpen() ? ' active' : '') + '" id="btnFilePreviewSearch">' + (fileBrowserIsSearchOpen() ? '关闭搜索' : '搜索') + '</button>';
        }
        buttons += '<button type="button" class="btn btn-sm btn-primary" id="btnFilePreviewSave"' + ((!fileBrowserIsDirty() || state.savingPreview || (state.previewReadResult && state.previewReadResult.truncated)) ? ' disabled' : '') + '>' + (state.savingPreview ? '保存中...' : '保存') + '</button>';
    }

    actionsEl.innerHTML = buttons;

    var previewBtn = document.getElementById('btnFilePreviewModePreview');
    if (previewBtn) {
        previewBtn.addEventListener('click', function() {
            switchFilePreviewRenderMode('preview');
        });
    }
    var editBtn = document.getElementById('btnFilePreviewModeEdit');
    if (editBtn) {
        editBtn.addEventListener('click', function() {
            switchFilePreviewRenderMode('edit');
        });
    }
    var searchBtn = document.getElementById('btnFilePreviewSearch');
    if (searchBtn) {
        refreshFileBrowserSearchButtonState();
        searchBtn.addEventListener('click', function() {
            if (state.previewEditorInstance && window.ProjectConfigCodeEditor) {
                window.ProjectConfigCodeEditor.toggleSearch(state.previewEditorInstance);
                refreshFileBrowserSearchButtonState();
            }
        });
    }
    var saveBtn = document.getElementById('btnFilePreviewSave');
    if (saveBtn) {
        saveBtn.addEventListener('click', saveCurrentFilePreview);
    }
}

function renderFilePreviewEditor(item, meta, readData) {
    var state = window.fileBrowserState;
    var bodyEl = document.getElementById('filePreviewBody');
    if (!bodyEl) return;
    destroyFileBrowserEditor();
    var truncatedHint = readData && readData.truncated
        ? '<div class="file-browser-editor-hint error">当前仅加载前 2MB 内容，已禁止保存，请使用外部编辑器处理大文件。</div>':'';
    bodyEl.innerHTML = '<div class="file-browser-editor-wrap">' +
        truncatedHint +
        '<div class="file-browser-code-editor" id="fileBrowserEditor"></div>' +
    '</div>';
    var editorMount = document.getElementById('fileBrowserEditor');
    if (!editorMount || !window.ProjectConfigCodeEditor) return;
    state.previewEditorInstance = window.ProjectConfigCodeEditor.create(editorMount, {
        fileName: item.name || meta.name || '',
        content: state.previewEditorValue || '',
        isDark: isFileBrowserDarkTheme(),
        onChange: function(value) {
            state.previewEditorValue = value || '';
            renderFilePreviewToolbar();
        }
    });
    if ((state.previewEditorValue || '') === (state.previewOriginalContent || '')) {
        window.ProjectConfigCodeEditor.markClean(state.previewEditorInstance);
    }
    window.ProjectConfigCodeEditor.focus(state.previewEditorInstance);
    startFileBrowserSearchButtonSync();
    renderFilePreviewToolbar();
}

function renderTextualFilePreview(item, meta, readData, ext) {
    var state = window.fileBrowserState;
    var bodyEl = document.getElementById('filePreviewBody');
    if (!bodyEl) return;
    if (state.previewRenderMode === 'edit' && fileBrowserCanEdit(meta)) {
        renderFilePreviewEditor(item, meta, readData);
        return;
    }
    if (meta.previewKind === 'markdown') {
        bodyEl.innerHTML = '<div class="oc-text file-browser-markdown">' + fileBrowserSanitizeMarkedHtml(marked.parse(readData.content || '')) + '</div>';
        return;
    }
    if (meta.previewKind === 'csv') {
        bodyEl.innerHTML = renderCSVPreview(readData.content || '');
        return;
    }
    bodyEl.innerHTML = '<pre class="file-browser-code"><code class="hljs">' + fileBrowserHighlightCode(readData.content || '', ext) + '</code></pre>';
}

function switchFilePreviewRenderMode(mode) {
    var state = window.fileBrowserState;
    var meta = state.previewMeta;
    if (!meta) return;
    if (mode === 'edit' && !fileBrowserCanEdit(meta)) return;
    if (mode === 'preview' && !fileBrowserCanPreview(meta)) return;
    if (mode === 'preview') {
        destroyFileBrowserEditor();
    }
    state.previewRenderMode = mode === 'edit' ? 'edit' : 'preview';
    renderFilePreviewToolbar();
    if (state.selectedItem) {
        renderFilePreview(state.selectedItem, { keepMode: true, skipMetaReload: true });
    }
}

async function saveCurrentFilePreview() {
    var state = window.fileBrowserState;
    var meta = state.previewMeta;
    var item = state.selectedItem;
    if (!item || !meta || !fileBrowserCanEdit(meta) || state.savingPreview) return;
    if (state.previewReadResult && state.previewReadResult.truncated) {
        showToast('当前文件已截断，禁止保存', 'error');
        return;
    }
    state.savingPreview = true;
    renderFilePreviewToolbar();
    try {
        var result = await fileBrowserApiSave(state.rootDir, item.path, state.previewEditorValue || '');
        if (!result.success) {
            showToast(result.error || '保存失败', 'error');
            return;
        }
        state.previewOriginalContent = state.previewEditorValue || '';
        state.previewContent = state.previewEditorValue || '';
        if (state.previewEditorInstance && window.ProjectConfigCodeEditor) {
            window.ProjectConfigCodeEditor.markClean(state.previewEditorInstance);
        }
        showToast('保存成功', 'success');
        renderFilePreviewToolbar();
        await loadFileBrowserList(state.currentPath || '/');
        state.selectedItem = state.items.find(function(entry) { return entry.path === item.path; }) || item;
        renderFileBrowserSelection();
        await renderFilePreview(state.selectedItem, { keepMode: true });
    } catch (err) {
        showToast(err.message || '保存失败', 'error');
    } finally {
        state.savingPreview = false;
        renderFilePreviewToolbar();
    }
}

async function renderFilePreview(item, options) {
    var state = window.fileBrowserState;
    options = options || {};
    if (!state || !item || item.type !== 'file') return;
    if (!(options.keepMode && state.previewRenderMode === 'edit')) {
        destroyFileBrowserEditor();
    }
    fileBrowserClearObjectURL();
    state.selectedItem = item;
    state.previewMode = 'file';
    updateFileBrowserDownloadButton(item);
    var titleEl = document.getElementById('filePreviewTitle');
    var metaEl = document.getElementById('filePreviewMeta');
    var bodyEl = document.getElementById('filePreviewBody');
    if (titleEl) titleEl.textContent = item.name;
    if (metaEl) metaEl.textContent = '加载中...';
    if (bodyEl) bodyEl.innerHTML = '<div class="file-browser-empty">正在读取文件...</div>';
    renderFilePreviewToolbar();

    try {
        var meta = options.skipMetaReload && state.previewMeta ? state.previewMeta : await fileBrowserApiStat(state.rootDir, item.path);
        state.previewMeta = meta;
        if (!options.keepMode) {
            state.previewRenderMode = fileBrowserGetPreferredRenderMode(meta);
        }
        if (metaEl) {
            metaEl.textContent = [meta.ext || '', fileBrowserFormatBytes(meta.size || 0), meta.modifiedAt || ''].filter(Boolean).join(' · ');
        }
        renderFilePreviewToolbar();
        var ext = (meta.ext || '').toLowerCase();
        var previewKind = meta.previewKind || '';

        if (previewKind === 'image') {
            var previewImageRes = await fileBrowserResolveRawResource(state.rootDir, item.path);
            bodyEl.innerHTML = '<div class="file-browser-image-wrap"><img class="file-browser-image" src="' + previewImageRes.url + '" alt="' + fileBrowserEscapeHTML(item.name) + '"></div>';
            return;
        }
        if (previewKind === 'pdf') {
            var previewPdfRes = await fileBrowserResolveRawResource(state.rootDir, item.path);
            bodyEl.innerHTML = '<iframe class="file-browser-pdf" title="PDF预览" src="' + previewPdfRes.url + '"></iframe>';
            return;
        }
        if (previewKind === 'spreadsheet') {
            bodyEl.innerHTML = '<div class="file-browser-unsupported">' +
                '<p>当前版本未启用 Excel 在线预览。</p>' +
                '<p>文件：' + fileBrowserEscapeHTML(item.name) + '</p>' +
                '</div>';
            return;
        }
        if (previewKind === 'markdown' || previewKind === 'csv' || previewKind === 'text' || previewKind === 'code') {
            var previewReadData = await fileBrowserApiRead(state.rootDir, item.path);
            state.previewReadResult = previewReadData;
            state.previewContent = previewReadData.content || '';
            if (!options.keepMode || state.previewEditorValue === '' || !fileBrowserIsDirty()) {
                state.previewEditorValue = previewReadData.content || '';
                state.previewOriginalContent = previewReadData.content || '';
            }
            renderTextualFilePreview(item, meta, previewReadData, ext);
            renderFilePreviewToolbar();
            return;
        }

        if (!ext) {
            if (state.previewRenderMode === 'edit' || state.forcedTextPreview[item.path]) {
                var noExtReadData = await fileBrowserApiRead(state.rootDir, item.path);
                var noExtContent = noExtReadData.content || '';
                state.previewReadResult = noExtReadData;
                state.previewContent = noExtContent;
                if (!options.keepMode || state.previewEditorValue === '' || !fileBrowserIsDirty()) {
                    state.previewEditorValue = noExtContent;
                    state.previewOriginalContent = noExtContent;
                }
                if (metaEl) {
                    metaEl.textContent = ['无扩展名 · 按普通文本方式打开', fileBrowserFormatBytes(meta.size || 0), meta.modifiedAt || ''].filter(Boolean).join(' · ');
                }
                renderTextualFilePreview(item, meta, noExtReadData, '');
                renderFilePreviewToolbar();
                return;
            }
            bodyEl.innerHTML = renderNoExtPreview(item, meta);
            bindNoExtPreviewActions(item);
            renderFilePreviewToolbar();
            return;
        }

        bodyEl.innerHTML = '<div class="file-browser-unsupported">' +
            '<p>该文件类型暂不支持在线预览。</p>' +
            '<p>文件：' + fileBrowserEscapeHTML(item.name) + '</p>' +
            '</div>';
        renderFilePreviewToolbar();
    } catch (err) {
        if (metaEl) metaEl.textContent = '';
        if (bodyEl) bodyEl.innerHTML = '<div class="file-browser-empty error">' + fileBrowserEscapeHTML(err.message || err) + '</div>';
        renderFilePreviewToolbar();
    }
}

function renderNoExtPreview(item, meta) {
    return '<div class="file-browser-noext">' +
        '<div class="file-browser-noext-title">无扩展名文件</div>' +
        '<div class="file-browser-noext-hint">系统暂时无法自动判断该文件类型。你可以按普通文本方式尝试预览。</div>' +
        '<div class="file-browser-noext-name">文件：' + fileBrowserEscapeHTML(item.name || meta.name || item.path || '') + '</div>' +
        '<div class="file-browser-noext-actions">' +
            '<button type="button" class="btn btn-sm btn-refresh" id="btnOpenNoExtAsText">按普通文本打开</button>' +
        '</div>' +
    '</div>';
}

function bindNoExtPreviewActions(item) {
    var openBtn = document.getElementById('btnOpenNoExtAsText');
    if (openBtn) {
        openBtn.onclick = function() {
            window.fileBrowserState.forcedTextPreview[item.path] = true;
            window.fileBrowserState.previewRenderMode = 'edit';
            renderFilePreview(item);
        };
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

async function renderGitFilePreview(path) {
    var state = window.fileBrowserState;
    if (!state || !path) return;
    fileBrowserClearObjectURL();
    state.previewMode = 'git';
    state.selectedItem = null;
    updateFileBrowserDownloadButton(null);
    var titleEl = document.getElementById('filePreviewTitle');
    var metaEl = document.getElementById('filePreviewMeta');
    var bodyEl = document.getElementById('filePreviewBody');
    if (titleEl) titleEl.textContent = path.replace(/^\//, '');
    if (metaEl) metaEl.textContent = 'Git 变更预览';
    if (bodyEl) bodyEl.innerHTML = '<div class="file-browser-empty">正在读取 Git 变更...</div>';
    renderFilePreviewToolbar();
    try {
        var data = await fileBrowserApiGitPreview(state.rootDir, path);
        if (!data.tracked) {
            bodyEl.innerHTML = '<div class="git-preview-section">' +
                '<div class="git-preview-section-title">未跟踪文件</div>' +
                '<div class="file-browser-empty" style="padding:0 0 12px">当前文件尚未纳入 Git 管理，因此没有历史 diff。</div>' +
                '<pre class="file-browser-code"><code class="hljs">' + fileBrowserHighlightCode(data.untrackedContent || '', '.' + ((data.path || '').split('.').pop() || '').toLowerCase()) + '</code></pre>' +
                '</div>';
            return;
        }
        bodyEl.innerHTML = '' +
            renderGitSection('已暂存修改', data.stagedBlocks || [], data.hasStaged) +
            renderGitSection('未暂存修改', data.unstagedBlocks || [], data.hasUnstaged);
    } catch (err) {
        bodyEl.innerHTML = '<div class="file-browser-empty error">' + fileBrowserEscapeHTML(err.message || err) + '</div>';
    }
}

async function renderGitHistoryFilePreview(commitHash, path) {
    var state = window.fileBrowserState;
    if (!state || !commitHash || !path) return;
    fileBrowserClearObjectURL();
    state.previewMode = 'git-history';
    state.selectedItem = null;
    updateFileBrowserDownloadButton(null);
    var titleEl = document.getElementById('filePreviewTitle');
    var metaEl = document.getElementById('filePreviewMeta');
    var bodyEl = document.getElementById('filePreviewBody');
    if (titleEl) titleEl.textContent = path;
    if (metaEl) metaEl.textContent = '提交历史 · ' + commitHash.slice(0, 7);
    if (bodyEl) bodyEl.innerHTML = '<div class="file-browser-empty">正在读取提交历史文件变更...</div>';
    renderFilePreviewToolbar();
    try {
        var data = await fileBrowserApiGitHistoryPreview(state.rootDir, commitHash, path);
        var blocks = data.blocks || [];
        bodyEl.innerHTML = renderGitSection('提交修改', blocks, true);
    } catch (err) {
        bodyEl.innerHTML = '<div class="file-browser-empty error">' + fileBrowserEscapeHTML(err.message || err) + '</div>';
    }
}

function renderGitSection(title, blocks, enabled) {
    if (!enabled || !blocks || !blocks.length) {
        return '<div class="git-preview-section">' +
            '<div class="git-preview-section-title">' + fileBrowserEscapeHTML(title) + '</div>' +
            '<div class="file-browser-empty">当前没有' + fileBrowserEscapeHTML(title) + '</div>' +
        '</div>';
    }
    var html = '<div class="git-preview-section">' +
        '<div class="git-preview-section-title">' + fileBrowserEscapeHTML(title) + '</div>';
    blocks.forEach(function(block) {
        html += renderGitDiffBlock(block);
    });
    html += '</div>';
    return html;
}

function renderGitDiffBlock(block) {
    var leftLines = block.left || [];
    var rightLines = block.right || [];
    var maxLen = Math.max(leftLines.length, rightLines.length);
    var html = '<div class="git-diff-grid">';
    for (var i = 0; i < maxLen; i++) {
        var left = leftLines[i] || { kind: 'empty', oldNo: 0, newNo: 0, text: '' };
        var right = rightLines[i] || { kind: 'empty', oldNo: 0, newNo: 0, text: '' };
        html += renderGitDiffLine(left, 'left');
        html += renderGitDiffLine(right, 'right');
    }
    html += '</div>';
    return html;
}

function renderGitDiffLine(line, side) {
    var no = side === 'left' ? line.oldNo : line.newNo;
    var noText = no ? String(no) : '';
    return '<div class="git-diff-line ' + fileBrowserEscapeHTML(line.kind || 'context') + '">' +
        '<span class="git-diff-line-no">' + fileBrowserEscapeHTML(noText) + '</span>' +
        '<span class="git-diff-line-text">' + fileBrowserEscapeHTML(line.text || '') + '</span>' +
    '</div>';
}
