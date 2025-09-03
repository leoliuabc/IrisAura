import './style.css';
import './app.css';

import logo from './assets/images/logo.png';
import { SelectFolder, CompressImages, TestFunction, GetSupportedFormats } from '../wailsjs/go/main/App';

document.querySelector('#app').innerHTML = `
    <img id="logo" class="logo">
    <div class="controls">
        <div class="input-row">
            <input type="text" id="inputDir" placeholder="输入文件夹" readonly>
            <button id="selectInputBtn">选择输入文件夹</button>
        </div>
        <div class="input-row">
            <input type="text" id="outputDir" placeholder="输出文件夹" readonly>
            <button id="selectOutputBtn">选择输出文件夹</button>
        </div>
        <div class="input-row">
            <select id="format"></select>
        </div>
        <div class="range-row">
            <label for="quality">质量: <span id="qualityValue">80</span></label>
            <input type="range" id="quality" min="1" max="100" value="80">
        </div>
        <div class="input-row">
            <input type="number" id="maxWidth" placeholder="最大宽度(px)">
            <input type="number" id="maxHeight" placeholder="最大高度(px)">
        </div>
        <button id="compressBtn">开始压缩</button>
    </div>

    <div id="loading" class="loading" style="display:none;">正在压缩，请稍候...</div>
    <div id="progress" class="progress" style="display:none;">
        <div id="progressFill" class="progress-fill">0%</div>
    </div>
    <div id="result" class="result" style="display:none;"></div>
`;

document.getElementById('logo').src = logo;

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i];
}

async function init() {
    const formatSelect = document.getElementById('format');
    const supportedFormats = await GetSupportedFormats();
    supportedFormats.forEach(f => {
        const option = document.createElement('option');
        option.value = f;
        option.text = f.toUpperCase();
        formatSelect.appendChild(option);
    });

    const qualitySlider = document.getElementById('quality');
    const qualityValue = document.getElementById('qualityValue');
    qualitySlider.addEventListener('input', () => {
        qualityValue.innerText = qualitySlider.value;
    });

    document.getElementById('selectInputBtn').addEventListener('click', async () => {
        const folder = await SelectFolder();
        if (folder) document.getElementById('inputDir').value = folder;
    });
    document.getElementById('selectOutputBtn').addEventListener('click', async () => {
        const folder = await SelectFolder();
        if (folder) document.getElementById('outputDir').value = folder;
    });

    const compressBtn = document.getElementById('compressBtn');
    const loading = document.getElementById('loading');
    const progress = document.getElementById('progress');
    const progressFill = document.getElementById('progressFill');
    const resultDiv = document.getElementById('result');

    compressBtn.addEventListener('click', async () => {
        const inputDir = document.getElementById('inputDir').value;
        const outputDir = document.getElementById('outputDir').value;
        if (!inputDir || !outputDir) {
            alert('请选择输入和输出文件夹');
            return;
        }

        compressBtn.disabled = true;
        loading.style.display = 'block';
        progress.style.display = 'block';
        progressFill.style.width = '0%';
        progressFill.innerText = '0%';
        resultDiv.style.display = 'none';
        resultDiv.innerText = '';

        const request = {
            inputDir,
            outputDir,
            format: formatSelect.value,
            quality: parseInt(qualitySlider.value),
            maxWidth: parseInt(document.getElementById('maxWidth').value) || 0,
            maxHeight: parseInt(document.getElementById('maxHeight').value) || 0
        };

        try {
            const result = await CompressImages(request);

            let percent = 0;
            const interval = setInterval(() => {
                percent += 5;
                if (percent > 100) percent = 100;
                progressFill.style.width = percent + '%';
                progressFill.innerText = percent + '%';
                if (percent === 100) clearInterval(interval);
            }, 50);

            setTimeout(() => {
                loading.style.display = 'none';
                progress.style.display = 'none';
                resultDiv.style.display = 'block';

                let msg = `处理: ${result.processedCount}/${result.totalCount}\n` +
                          `原始大小: ${formatFileSize(result.originalSize)}\n` +
                          `压缩后大小: ${formatFileSize(result.compressedSize)}\n` +
                          `${result.message}`;

                if (result.errors && result.errors.length > 0) {
                    msg += '\n\n错误:\n' + result.errors.join('\n');
                }

                resultDiv.innerText = msg;
            }, 500);

        } catch (err) {
            loading.style.display = 'none';
            progress.style.display = 'none';
            alert('压缩失败: ' + err.message);
        } finally {
            compressBtn.disabled = false;
        }
    });

    try {
        const test = await TestFunction();
        console.log('后端测试:', test);
    } catch (err) {
        console.error('后端连接失败:', err);
    }
}

init();