import os
from PIL import Image, ImageDraw, ImageFont, ImageFilter

output_dir = "/home/afari/Projects/ns/playstore_assets"
os.makedirs(output_dir, exist_ok=True)

logo_path = "/home/afari/Projects/ns/frontend/nsapp/assets/images/logo.png"
logo2_path = "/home/afari/Projects/ns/frontend/nsapp/assets/images/logo2.png"
splash_path = "/home/afari/Projects/ns/frontend/nsapp/assets/images/splash_img.png"

# ==========================================
# 1. APP ICON: Exactly 512 x 512 px
# ==========================================
icon_512_path = os.path.join(output_dir, "app_icon_512x512.png")

# Open logo
logo_img = Image.open(logo_path).convert("RGBA")

# Create 512x512 canvas with brand background or clean padding
# Let's check logo background or fit logo into 512x512
icon_canvas = Image.new("RGBA", (512, 512), (255, 255, 255, 0)) # transparent or brand color

# Resize logo to fit nicely with padding (e.g. 450x450 or exact fit)
logo_resized = logo_img.resize((512, 512), Image.Resampling.LANCZOS)
icon_canvas.paste(logo_resized, (0, 0), logo_resized)
icon_canvas.save(icon_512_path, "PNG")
print("Saved App Icon (512x512):", icon_512_path)

# ==========================================
# 2. FEATURE GRAPHIC: Exactly 1024 x 500 px
# ==========================================
feature_path = os.path.join(output_dir, "feature_graphic_1024x500.png")

# Create 1024x500 canvas with modern premium gradient
w, h = 1024, 500
feature_img = Image.new("RGBA", (w, h), (15, 23, 42, 255)) # Dark navy
draw = ImageDraw.Draw(feature_img)

# Draw subtle diagonal gradient / decorative shapes
for y in range(h):
    # Gradient from #0f172a (15, 23, 42) to #1e293b (30, 41, 59)
    r = int(15 + (30 - 15) * (y / h))
    g = int(23 + (41 - 23) * (y / h))
    b = int(42 + (59 - 42) * (y / h))
    draw.line([(0, y), (w, y)], fill=(r, g, b, 255))

# Draw decorative soft glow circle in background
glow = Image.new("RGBA", (w, h), (0, 0, 0, 0))
glow_draw = ImageDraw.Draw(glow)
glow_draw.ellipse([w - 400, -100, w + 200, 400], fill=(56, 189, 248, 40)) # cyan glow
glow_draw.ellipse([-100, 200, 400, 600], fill=(99, 102, 241, 40)) # indigo glow
glow = glow.filter(ImageFilter.GaussianBlur(50))
feature_img = Image.alpha_composite(feature_img, glow)
draw = ImageDraw.Draw(feature_img)

# Paste Logo on Left/Center
logo_w, logo_h = 320, 320
logo_feature = logo_img.resize((logo_w, logo_h), Image.Resampling.LANCZOS)

# Position logo on the left or center
feature_img.paste(logo_feature, (80, (h - logo_h) // 2), logo_feature)

# Draw Title & Tagline Text on the Right
try:
    font_title = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 52)
    font_tag = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 24)
    font_sub = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", 18)
except:
    font_title = ImageFont.load_default()
    font_tag = ImageFont.load_default()
    font_sub = ImageFont.load_default()

# Text details
draw.text((450, 160), "Neighbor Service", fill=(255, 255, 255, 255), font=font_title)
draw.text((450, 240), "Connect, Hire & Deliver Services Locally", fill=(56, 189, 248, 255), font=font_tag)
draw.text((450, 290), "Fast • Reliable • Trusted Local Community", fill=(148, 163, 184, 255), font=font_sub)

# Save Feature Graphic
feature_img.convert("RGB").save(feature_path, "PNG", quality=95)
print("Saved Feature Graphic (1024x500):", feature_path)
