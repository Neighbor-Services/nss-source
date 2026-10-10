import argparse
import subprocess
import os
import sys

def resize_video(input_file, orientation):
    if not os.path.exists(input_file):
        print(f"Error: File '{input_file}' not found.")
        sys.exit(1)

    if orientation == 'portrait':
        width, height = 886, 1920
    else:
        width, height = 1920, 886

    name, ext = os.path.splitext(input_file)
    output_file = f"{name}_resized{ext}"

    # We use scale and pad to fit it perfectly into the required dimensions without stretching
    vf_filter = f"scale={width}:{height}:force_original_aspect_ratio=decrease,pad={width}:{height}:(ow-iw)/2:(oh-ih)/2"
    
    cmd = [
        "ffmpeg",
        "-i", input_file,
        "-vf", vf_filter,
        "-c:v", "libx264",
        "-crf", "26", # Lower quality/smaller file size
        "-preset", "medium",
        "-r", "30", # App Store Connect requires <= 30 fps
        "-c:a", "aac", # App Store Connect requires AAC audio
        "-b:a", "256k",
        "-movflags", "+faststart", # Optimizes for web streaming/upload
        "-y", # Overwrite output if exists
        output_file
    ]

    print(f"Resizing {input_file} to {width}x{height}...")
    print(f"Command: {' '.join(cmd)}")
    
    try:
        subprocess.run(cmd, check=True)
        print(f"\n✅ Success! Resized video saved to: {output_file}")
    except FileNotFoundError:
        print("\n❌ Error: 'ffmpeg' is not installed. Please install it to use this script.")
        print("On macOS: brew install ffmpeg")
        print("On Linux: sudo apt install ffmpeg")
    except subprocess.CalledProcessError as e:
        print(f"\n❌ Error: FFmpeg failed with error code {e.returncode}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Resize video for Apple App Store Connect")
    parser.add_argument("input_file", help="Path to the input video file")
    parser.add_argument("--orientation", choices=['portrait', 'landscape'], default='portrait', 
                        help="Target orientation: 'portrait' (886x1920) or 'landscape' (1920x886). Default is portrait.")
    
    args = parser.parse_args()
    resize_video(args.input_file, args.orientation)
