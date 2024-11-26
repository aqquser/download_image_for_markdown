# 获取当前脚本所在目录
$markdownDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# 获取默认图片保存路径
$defaultSaveDir = Join-Path $env:USERPROFILE "Pictures\markdown图片"

# 获取用户输入的保存路径
Write-Host "请输入图片保存路径 (直接回车使用默认路径: $defaultSaveDir):"
$saveDir = Read-Host
if ([string]::IsNullOrEmpty($saveDir)) {
    $saveDir = $defaultSaveDir
}

# 检查并创建保存目录
if (-not (Test-Path $saveDir)) {
    Write-Host "路径 $saveDir 不存在，是否创建? (y/n):"
    $answer = Read-Host
    if ($answer.ToLower() -eq 'y') {
        New-Item -ItemType Directory -Path $saveDir -Force | Out-Null
    } else {
        Write-Host "程序退出"
        Read-Host "按回车键退出..."
        exit
    }
}

# 下载图片函数
function Download-Image {
    param (
        [string]$url,
        [string]$savePath
    )
    
    try {
        $webClient = New-Object System.Net.WebClient
        $webClient.DownloadFile($url, $savePath)
        return $true
    } catch {
        Write-Host "下载图片失败 $url`: $_"
        return $false
    }
}

# 处理Markdown文件函数
function Process-MarkdownFile {
    param (
        [string]$filePath
    )
    
    Write-Host "处理文件: $filePath"
    $content = Get-Content -Path $filePath -Raw -Encoding UTF8
    $modified = $false
    $imageCount = 0
    $hasNoedgeaiImages = $false
    
    # 匹配图片标签的正则表达式
    $imgPattern = '<img src="(.*?)"/>'
    
    # 添加调试信息：显示找到的所有匹配
    $matches = [regex]::Matches($content, $imgPattern)
    Write-Host "找到 $($matches.Count) 个图片标签"
    
    $newContent = $content
    foreach ($match in $matches) {
        $imageCount++
        $originalURL = $match.Groups[1].Value
        Write-Host "处理第 $imageCount 个图片: $originalURL"
        
        # 检查URL是否包含noedgeai.com
        if ($originalURL -like "*noedgeai.com*") {
            $hasNoedgeaiImages = $true
            
            # 获取文件扩展名
            $fileExt = [System.IO.Path]::GetExtension($originalURL.Split('?')[0])
            if ([string]::IsNullOrEmpty($fileExt)) {
                $fileExt = ".jpg"
            }
            
            # 构建新文件名
            $baseName = [System.IO.Path]::GetFileNameWithoutExtension($filePath)
            $newFilename = "${baseName}-${imageCount}${fileExt}"
            $savePath = Join-Path $saveDir $newFilename
            
            # 下载图片
            if (Download-Image -url $originalURL -savePath $savePath) {
                $modified = $true
                $newContent = $newContent -replace [regex]::Escape($match.Value), "<img src=`"$savePath`"/>"
                Write-Host "已替换图片路径: $savePath"
            }
        }
    }
    
    if ($modified) {
        Write-Host "准备写入更新后的内容到文件..."
        # 使用 UTF-8 编码保存文件
        $Utf8NoBomEncoding = New-Object System.Text.UTF8Encoding $False
        [System.IO.File]::WriteAllText($filePath, $newContent, $Utf8NoBomEncoding)
        Write-Host "文件已更新"
    } else {
        if (-not $hasNoedgeaiImages) {
            Write-Host "所有图片不包含noedgeai.com,文件无需更新"
        } else {
            Write-Host "文件无需更新"
        }
    }
}

# 处理目录下所有的markdown文件
Get-ChildItem -Path $markdownDir -Filter "*.md" -File | ForEach-Object {
    Process-MarkdownFile -filePath $_.FullName
}

Write-Host "`n处理完成！"
Read-Host "按回车键退出..."
