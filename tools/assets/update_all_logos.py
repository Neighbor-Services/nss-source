#!/usr/bin/env python3
"""
Comprehensive asset generator for Neighbor Services.
Updates:
- Flutter app assets (logo.png, logo2.png)
- Android launcher icons (legacy mipmaps & drawables)
- Android adaptive icons (API 26+) with safe-zone foreground & white background
- Android notification icons (monochrome white with alpha channel per Android guidelines)
- iOS AppIcon set (all required resolutions, RGB without alpha per Apple HIG)
- Google Play Store assets (512x512 app icon & feature graphic)
- Web portal assets (admin-portal & public-site logos and favicons)
"""

import os
import shutil
import numpy as np
from PIL import Image, ImageDraw, ImageFilter, ImageFont

SRC_IMG_PATH = "/home/afari/.gemini/antigravity-ide/brain/6fe7a14d-27ae-48b4-a592-503a55577018/.user_uploaded/media_1792641542382.png"
REPO_ROOT = "/home/afari/Projects/ns"
NSAPP_DIR = os.path.join(REPO_ROOT, "frontend/nsapp")
RES_DIR = os.path.join(NSAPP_DIR, "android/app/src/main/res")
IOS_APPICON_DIR = os.path.join(NSAPP_DIR, "ios/Runner/Assets.xcassets/AppIcon.appiconset")

def load_and_center_badge(src_path):
    src = Image.open(src_path).convert("RGBA")
    arr = np.array(src)
    
    # Mathematical circle fit
    # Fitted circle: cx=511.63, cy=509.46, radius=509.38
    cx, cy, r = 511.63, 509.46, 508.5
    dx = int(round(512 - cx))
    dy = int(round(512 - cy))
    
    shifted = Image.new("RGBA", (1024, 1024), (255, 255, 255, 0))
    shifted.paste(src, (dx, dy))
    
    # Anti-aliased alpha mask outside radius r
    w, h = 1024, 1024
    y, x = np.ogrid[:h, :w]
    dist = np.sqrt((x - 512.0)**2 + (y - 512.0)**2)
    edge_width = 2.0
    alpha = np.clip((r - dist) / edge_width + 0.5, 0.0, 1.0)
    
    arr_shifted = np.array(shifted)
    arr_shifted[:, :, 3] = (alpha * 255).astype(np.uint8)
    badge_clean = Image.fromarray(arr_shifted)
    return badge_clean

def generate_notification_silhouette(src_path):
    """
    Extracts the NS house monogram in pure white with alpha channel.
    Android notification icons MUST be monochrome white with alpha.
    """
    src = Image.open(src_path).convert("RGB")
    arr = np.array(src)
    
    # Monogram region: y in [210, 635], x in [235, 825]
    mono_crop = arr[210:635, 235:825]
    # White background is > 235 in all channels
    mask = ((mono_crop[:,:,0] < 235) | (mono_crop[:,:,1] < 235) | (mono_crop[:,:,2] < 235)).astype(np.float32)
    im_mask = Image.fromarray((mask * 255).astype(np.uint8))
    return im_mask

def create_notification_icon(mono_mask, size):
    """Generates a crisp notification icon at `size` x `size` in pure white with alpha."""
    content_size = int(size * 0.82)
    aspect = mono_mask.size[0] / mono_mask.size[1]
    if aspect > 1:
        cw = content_size
        ch = max(1, int(cw / aspect))
    else:
        ch = content_size
        cw = max(1, int(ch * aspect))
    
    scaled_mask = mono_mask.resize((cw, ch), Image.Resampling.LANCZOS)
    canvas_mask = Image.new("L", (size, size), 0)
    ox = (size - cw) // 2
    oy = (size - ch) // 2
    canvas_mask.paste(scaled_mask, (ox, oy))
    
    notif = Image.new("RGBA", (size, size), (255, 255, 255, 0))
    notif.putalpha(canvas_mask)
    return notif

def main():
    print("Loading badge from:", SRC_IMG_PATH)
    badge_rgba = load_and_center_badge(SRC_IMG_PATH)
    mono_mask = generate_notification_silhouette(SRC_IMG_PATH)
    
    # -------------------------------------------------------------
    # 1. Flutter in-app assets
    # -------------------------------------------------------------
    assets_img_dir = os.path.join(NSAPP_DIR, "assets/images")
    os.makedirs(assets_img_dir, exist_ok=True)
    
    # logo.png and logo2.png
    badge_rgba.save(os.path.join(assets_img_dir, "logo.png"), "PNG")
    badge_rgba.save(os.path.join(assets_img_dir, "logo2.png"), "PNG")
    print("Saved Flutter assets: logo.png, logo2.png")
    
    # -------------------------------------------------------------
    # 2. Android Launcher Icons (Legacy & Adaptive)
    # -------------------------------------------------------------
    # Densities: (folder, launcher_size, adaptive_fg_size)
    android_densities = [
        ("mipmap-mdpi", 48, 108),
        ("mipmap-hdpi", 72, 162),
        ("mipmap-xhdpi", 96, 216),
        ("mipmap-xxhdpi", 144, 324),
        ("mipmap-xxxhdpi", 192, 432),
    ]
    
    # For legacy mipmap launcher icons:
    # A circular badge with transparent corners, scaled to 88% so it floats nicely
    for folder, size, fg_size in android_densities:
        folder_path = os.path.join(RES_DIR, folder)
        os.makedirs(folder_path, exist_ok=True)
        
        # Legacy launcher icon (size x size)
        scale = 0.88
        scaled_badge = badge_rgba.resize((int(size * scale), int(size * scale)), Image.Resampling.LANCZOS)
        legacy_icon = Image.new("RGBA", (size, size), (0, 0, 0, 0))
        offset = (size - scaled_badge.size[0]) // 2
        legacy_icon.paste(scaled_badge, (offset, offset), scaled_badge)
        
        legacy_icon.save(os.path.join(folder_path, "ic_launcher.png"), "PNG")
        legacy_icon.save(os.path.join(folder_path, "logo.png"), "PNG")
        legacy_icon.save(os.path.join(folder_path, "logo2.png"), "PNG")
        
        # Adaptive foreground icon (fg_size x fg_size)
        # Safe zone diameter is 66-72dp out of 108dp (~68% of fg_size)
        fg_scale = 0.68
        badge_fg = badge_rgba.resize((int(fg_size * fg_scale), int(fg_size * fg_scale)), Image.Resampling.LANCZOS)
        fg_icon = Image.new("RGBA", (fg_size, fg_size), (0, 0, 0, 0))
        fg_offset = (fg_size - badge_fg.size[0]) // 2
        fg_icon.paste(badge_fg, (fg_offset, fg_offset), badge_fg)
        
        fg_icon.save(os.path.join(folder_path, "ic_launcher_foreground.png"), "PNG")
        fg_icon.save(os.path.join(folder_path, "logo_foreground.png"), "PNG")
    
    # Also update drawable & drawable-v21
    for d_folder in ["drawable", "drawable-v21"]:
        d_path = os.path.join(RES_DIR, d_folder)
        os.makedirs(d_path, exist_ok=True)
        badge_500 = badge_rgba.resize((500, 500), Image.Resampling.LANCZOS)
        badge_500.save(os.path.join(d_path, "logo.png"), "PNG")
        badge_500.save(os.path.join(d_path, "logo2.png"), "PNG")
        badge_500.save(os.path.join(d_path, "ic_launcher.png"), "PNG")
        
    # Adaptive XMLs in mipmap-anydpi-v26
    v26_dir = os.path.join(RES_DIR, "mipmap-anydpi-v26")
    os.makedirs(v26_dir, exist_ok=True)
    adaptive_xml_content = """<?xml version="1.0" encoding="utf-8"?>
<adaptive-icon xmlns:android="http://schemas.android.com/apk/res/android">
    <background android:drawable="@color/ic_launcher_background"/>
    <foreground android:drawable="@mipmap/ic_launcher_foreground"/>
</adaptive-icon>
"""
    with open(os.path.join(v26_dir, "ic_launcher.xml"), "w") as f:
        f.write(adaptive_xml_content)
        
    logo_adaptive_xml = """<?xml version="1.0" encoding="utf-8"?>
<adaptive-icon xmlns:android="http://schemas.android.com/apk/res/android">
    <background android:drawable="@color/ic_launcher_background"/>
    <foreground android:drawable="@mipmap/logo_foreground"/>
</adaptive-icon>
"""
    with open(os.path.join(v26_dir, "logo.xml"), "w") as f:
        f.write(logo_adaptive_xml)
        
    # Update colors.xml with ic_launcher_background
    colors_xml_path = os.path.join(RES_DIR, "values/colors.xml")
    with open(colors_xml_path, "w") as f:
        f.write("""<?xml version="1.0" encoding="utf-8"?>
<resources>
    <color name="notification_color">#5E17EB</color>
    <color name="ic_launcher_background">#FFFFFF</color>
</resources>
""")
    print("Updated Android launcher & adaptive icons.")

    # -------------------------------------------------------------
    # 3. Android Notification Icons
    # -------------------------------------------------------------
    # Notification sizes: mdpi: 24, hdpi: 36, xhdpi: 48, xxhdpi: 72, xxxhdpi: 96
    notif_sizes = {
        "drawable-mdpi": 24,
        "drawable-hdpi": 36,
        "drawable-xhdpi": 48,
        "drawable-xxhdpi": 72,
        "drawable-xxxhdpi": 96,
        "drawable": 96,
        "drawable-v21": 96,
        "mipmap-mdpi": 24,
        "mipmap-hdpi": 36,
        "mipmap-xhdpi": 48,
        "mipmap-xxhdpi": 72,
        "mipmap-xxxhdpi": 96,
    }
    for folder, size in notif_sizes.items():
        folder_path = os.path.join(RES_DIR, folder)
        os.makedirs(folder_path, exist_ok=True)
        notif_img = create_notification_icon(mono_mask, size)
        notif_img.save(os.path.join(folder_path, "ic_notification.png"), "PNG")
    print("Updated Android notification icons across all drawable and mipmap folders.")

    # -------------------------------------------------------------
    # 4. iOS AppIcon Set
    # -------------------------------------------------------------
    # Apple HIG: RGB (no alpha), centered on white background with ~14% margin
    ios_scale = 0.82
    badge_ios = badge_rgba.resize((int(1024 * ios_scale), int(1024 * ios_scale)), Image.Resampling.LANCZOS)
    ios_base = Image.new("RGB", (1024, 1024), (255, 255, 255))
    offset = (1024 - badge_ios.size[0]) // 2
    ios_base.paste(badge_ios, (offset, offset), badge_ios)
    
    ios_sizes = [
        ("20.png", 20),
        ("29.png", 29),
        ("40.png", 40),
        ("57.png", 57),
        ("58.png", 58),
        ("60.png", 60),
        ("76.png", 76),
        ("80.png", 80),
        ("87.png", 87),
        ("114.png", 114),
        ("120.png", 120),
        ("152.png", 152),
        ("167.png", 167),
        ("167_ipad.png", 167),
        ("180.png", 180),
        ("1024.png", 1024),
    ]
    os.makedirs(IOS_APPICON_DIR, exist_ok=True)
    for filename, size in ios_sizes:
        resized_ios = ios_base.resize((size, size), Image.Resampling.LANCZOS)
        resized_ios.save(os.path.join(IOS_APPICON_DIR, filename), "PNG")
    print("Updated iOS AppIcon set (16 resolutions, RGB without alpha).")

    # -------------------------------------------------------------
    # 5. Google Play Store Assets
    # -------------------------------------------------------------
    playstore_dir = os.path.join(REPO_ROOT, "playstore_assets")
    os.makedirs(playstore_dir, exist_ok=True)
    
    # 512x512 app icon
    badge_512 = badge_rgba.resize((int(512 * 0.88), int(512 * 0.88)), Image.Resampling.LANCZOS)
    play_icon = Image.new("RGBA", (512, 512), (255, 255, 255, 0))
    offset_512 = (512 - badge_512.size[0]) // 2
    play_icon.paste(badge_512, (offset_512, offset_512), badge_512)
    play_icon.save(os.path.join(playstore_dir, "app_icon_512x512.png"), "PNG")
    
    # Feature Graphic (1024x500)
    w_feat, h_feat = 1024, 500
    feat_img = Image.new("RGBA", (w_feat, h_feat), (15, 23, 42, 255))
    draw = ImageDraw.Draw(feat_img)
    for y_idx in range(h_feat):
        r_c = int(15 + (30 - 15) * (y_idx / h_feat))
        g_c = int(23 + (41 - 23) * (y_idx / h_feat))
        b_c = int(42 + (59 - 42) * (y_idx / h_feat))
        draw.line([(0, y_idx), (w_feat, y_idx)], fill=(r_c, g_c, b_c, 255))
        
    glow = Image.new("RGBA", (w_feat, h_feat), (0, 0, 0, 0))
    glow_draw = ImageDraw.Draw(glow)
    glow_draw.ellipse([w_feat - 400, -100, w_feat + 200, 400], fill=(56, 189, 248, 40))
    glow_draw.ellipse([-100, 200, 400, 600], fill=(99, 102, 241, 40))
    glow = glow.filter(ImageFilter.GaussianBlur(50))
    feat_img = Image.alpha_composite(feat_img, glow)
    
    logo_w, logo_h = 320, 320
    badge_feat = badge_rgba.resize((logo_w, logo_h), Image.Resampling.LANCZOS)
    feat_img.paste(badge_feat, (80, (h_feat - logo_h) // 2), badge_feat)
    
    draw_f = ImageDraw.Draw(feat_img)
    try:
        font_title = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 52)
        font_tag = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 24)
        font_sub = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 18)
    except:
        font_title = ImageFont.load_default()
        font_tag = ImageFont.load_default()
        font_sub = ImageFont.load_default()
        
    draw_f.text((450, 160), "Neighbor Services", fill=(255, 255, 255, 255), font=font_title)
    draw_f.text((450, 240), "Connect, Hire & Deliver Services Locally", fill=(56, 189, 248, 255), font=font_tag)
    draw_f.text((450, 290), "Neighbors Helping Neighbors", fill=(148, 163, 184, 255), font=font_sub)
    feat_img.convert("RGB").save(os.path.join(playstore_dir, "feature_graphic_1024x500.png"), "PNG")
    print("Updated Play Store assets.")

    # -------------------------------------------------------------
    # 6. Web Portals (admin-portal & public-site)
    # -------------------------------------------------------------
    for web_dir in ["frontend/admin-portal", "frontend/public-site"]:
        target_img_dir = os.path.join(REPO_ROOT, web_dir, "public/images")
        if os.path.exists(target_img_dir):
            badge_500 = badge_rgba.resize((500, 500), Image.Resampling.LANCZOS)
            badge_500.save(os.path.join(target_img_dir, "logo.png"), "PNG")
            
        target_favicon = os.path.join(REPO_ROOT, web_dir, "public/favicon.ico")
        if os.path.exists(os.path.dirname(target_favicon)):
            badge_fav = badge_rgba.resize((64, 64), Image.Resampling.LANCZOS)
            badge_fav.save(target_favicon, format="ICO", sizes=[(16, 16), (32, 32), (48, 48), (64, 64)])
    print("Updated web portal logos and favicons.")

if __name__ == "__main__":
    main()
