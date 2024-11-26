# 编译命令  
# pyinstaller --onefile --clean --noconfirm  --name "Doc2X_img_to_local_python" Doc2X_img_to_local.py
import os
import re
import requests
import tempfile
import subprocess
import json
from pathlib import Path
import sys

class ImageDownloader:
    def __init__(self, save_dir):
        self.save_dir = save_dir
        
    def download_image(self, url, save_path):
        """下载图片到指定路径"""
        try:
            response = requests.get(url, timeout=30)
            if response.status_code == 200:
                with open(save_path, 'wb') as f:
                    f.write(response.content)
                return True
        except Exception as e:
            print(f"下载图片失败 {url}: {e}")
        return False

    def process_markdown_file(self, file_path):
        """处理单个markdown文件"""
        print(f"处理文件: {file_path}")
        
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
        except Exception as e:
            print(f"读取文件失败 {file_path}: {e}")
            return
        
        # 匹配图片标签
        img_pattern = r'<img src="(.*?)"/>'
        modified = False
        image_count = 0
        no_update_printed = False  # 新增标志变量
        
        def replace_image(match):
            nonlocal modified, image_count, no_update_printed
            original_url = match.group(1)
            image_count += 1
            
            # 检查图片地址是否包含"noedgeai.com"
            if "noedgeai.com" not in original_url:
                if not no_update_printed:  # 只输出一次提示
                    print(f"图片不包含noedgeai.com,文件无需更新")
                    no_update_printed = True
                return match.group(0)
            
            # 获取原始文件扩展名
            file_ext = os.path.splitext(original_url.split('?')[0])[1] or '.jpg'
            
            # 构建新的文件名
            base_name = os.path.splitext(os.path.basename(file_path))[0]
            new_filename = f"{base_name}-{image_count}{file_ext}"
            save_path = os.path.join(self.save_dir, new_filename)
            
            # 下载图片
            if self.download_image(original_url, save_path):
                modified = True
                # 使用相对路径替换URL
                return f'<img src="{save_path}"/>'
            
            return match.group(0)
        
        new_content = re.sub(img_pattern, replace_image, content)
        
        # 只有当内容被修改时才写入文件
        if modified:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(new_content)
            print(f"文件已更新")
        else:
            if not no_update_printed:  # 确保在没有修改时也输出一次提示
                print(f"文件无需更新")

def main():
    # 获取程序所在目录
    exec_path = Path(sys.executable if getattr(sys, 'frozen', False) else __file__)
    markdown_dir = exec_path.parent
    
    # 获取用户主目录和默认保存路径
    try:
        current_user = Path.home()
        default_save_dir = current_user / "Pictures" / "markdown图片"
    except Exception as e:
        print(f"获取用户目录失败: {e}")
        input("\n按回车键退出...")
        return
    
    # 获取用户输入的保存路径
    print(f"请输入图片保存路径 (直接回车使用默认路径: {default_save_dir}): ")
    save_dir = input().strip()
    
    # 如果用户直接按回车，使用默认路径
    if not save_dir:
        save_dir = str(default_save_dir)
    
    # 验证路径是否存在
    save_path = Path(save_dir)
    if not save_path.exists():
        answer = input(f"路径 {save_dir} 不存在，是否创建? (y/n): ").lower()
        if answer == 'y':
            try:
                save_path.mkdir(parents=True, exist_ok=True)
            except Exception as e:
                print(f"创建目录失败: {e}")
                input("\n按回车键退出...")
                return
        else:
            print("程序退出")
            input("\n按回车键退出...")
            return

    downloader = ImageDownloader(str(save_path))
    
    # 遍历目录中的所有markdown文件
    try:
        for file_path in markdown_dir.rglob("*.md"):
            downloader.process_markdown_file(str(file_path))
    except Exception as e:
        print(f"遍历目录失败: {e}")
    
    print("\n处理完成！按回车键退出...")
    input()

if __name__ == "__main__":
    main()
