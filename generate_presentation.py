import os
from pptx import Presentation
from pptx.util import Inches, Pt
from pptx.dml.color import RGBColor
from pptx.enum.text import PP_ALIGN, MSO_ANCHOR
from pptx.enum.shapes import MSO_SHAPE

def create_deck(output_path):
    prs = Presentation()
    # 16:9 widescreen slides (13.333" x 7.5")
    prs.slide_width = Inches(13.333)
    prs.slide_height = Inches(7.5)
    blank_layout = prs.slide_layouts[6]

    # --- Modern Dark UI Design Palette ---
    BG_MAIN         = RGBColor(11, 17, 32)      # Deep Obsidian Navy (#0b1120)
    CARD_BG         = RGBColor(20, 30, 52)      # Rich Slate Navy (#141e34)
    CARD_BORDER     = RGBColor(37, 54, 88)      # Crisp Border (#253658)
    CARD_HEADER_BG  = RGBColor(28, 42, 72)      # Sub-card / Header background
    
    ACCENT_CYAN     = RGBColor(34, 211, 238)    # Sky / Cyan (#22d3ee)
    ACCENT_BLUE     = RGBColor(96, 165, 250)    # Electric Blue (#60a5fa)
    ACCENT_EMERALD  = RGBColor(52, 211, 153)    # Success Emerald (#34d399)
    ACCENT_AMBER    = RGBColor(251, 191, 36)    # Warm Amber (#fbbf24)
    ACCENT_ROSE     = RGBColor(251, 113, 133)   # Danger Rose (#fb7185)
    ACCENT_PURPLE   = RGBColor(192, 132, 252)   # Violet (#c084fc)
    
    TEXT_WHITE      = RGBColor(248, 250, 252)   # Slate 50
    TEXT_BODY       = RGBColor(226, 232, 240)   # Slate 200
    TEXT_MUTED      = RGBColor(148, 163, 184)   # Slate 400
    TEXT_DIM        = RGBColor(100, 116, 139)   # Slate 500

    FONT_MAIN = "Segoe UI"

    def apply_slide_bg(slide):
        bg = slide.shapes.add_shape(MSO_SHAPE.RECTANGLE, 0, 0, prs.slide_width, prs.slide_height)
        bg.fill.solid()
        bg.fill.fore_color.rgb = BG_MAIN
        bg.line.fill.background()
        return bg

    def add_header(slide, title, category, slide_num, total_slides=22):
        # Category Pill
        pill = slide.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, Inches(0.8), Inches(0.42), Inches(3.2), Inches(0.32))
        pill.fill.solid()
        pill.fill.fore_color.rgb = CARD_HEADER_BG
        pill.line.color.rgb = CARD_BORDER
        pill.line.width = Pt(1)
        
        ptf = pill.text_frame
        ptf.margin_top = Inches(0)
        ptf.margin_left = Inches(0)
        ptf.margin_right = Inches(0)
        ptf.margin_bottom = Inches(0)
        pp = ptf.paragraphs[0]
        pp.alignment = PP_ALIGN.CENTER
        pp.text = category.upper()
        pp.font.name = FONT_MAIN
        pp.font.size = Pt(9.5)
        pp.font.bold = True
        pp.font.color.rgb = ACCENT_CYAN

        # Main Title
        tb = slide.shapes.add_textbox(Inches(0.8), Inches(0.8), Inches(10.5), Inches(0.65))
        tf = tb.text_frame
        tf.word_wrap = True
        tf.margin_top = Inches(0)
        tf.margin_left = Inches(0)
        p = tf.paragraphs[0]
        p.text = title
        p.font.name = FONT_MAIN
        p.font.size = Pt(22)
        p.font.bold = True
        p.font.color.rgb = TEXT_WHITE

        # Top Accent Divider
        div = slide.shapes.add_shape(MSO_SHAPE.RECTANGLE, Inches(0.8), Inches(1.5), Inches(11.73), Inches(0.02))
        div.fill.solid()
        div.fill.fore_color.rgb = CARD_BORDER
        div.line.fill.background()

        # Footer branding & slide number
        foot_div = slide.shapes.add_shape(MSO_SHAPE.RECTANGLE, Inches(0.8), Inches(7.05), Inches(11.73), Inches(0.015))
        foot_div.fill.solid()
        foot_div.fill.fore_color.rgb = CARD_BORDER
        foot_div.line.fill.background()

        ftb = slide.shapes.add_textbox(Inches(0.8), Inches(7.1), Inches(8.0), Inches(0.3))
        ftf = ftb.text_frame
        fp = ftf.paragraphs[0]
        fp.text = "High-Performance Multi-Server Architecture  |  Contabo Infrastructure Demo"
        fp.font.name = FONT_MAIN
        fp.font.size = Pt(9)
        fp.font.color.rgb = TEXT_DIM

        # Slide Number
        sntb = slide.shapes.add_textbox(Inches(10.5), Inches(7.1), Inches(2.0), Inches(0.3))
        sntf = sntb.text_frame
        snp = sntf.paragraphs[0]
        snp.alignment = PP_ALIGN.RIGHT
        snp.text = f"{slide_num:02d} / {total_slides:02d}"
        snp.font.name = FONT_MAIN
        snp.font.size = Pt(9.5)
        snp.font.bold = True
        snp.font.color.rgb = ACCENT_CYAN

    def add_pro_card(slide, left, top, width, height, title, points, accent_color=ACCENT_BLUE, badge_text=None, subtitle=None):
        # Card Body
        card = slide.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, left, top, width, height)
        card.fill.solid()
        card.fill.fore_color.rgb = CARD_BG
        card.line.color.rgb = CARD_BORDER
        card.line.width = Pt(1)

        # Top Colored Accent Stripe
        stripe = slide.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, left, top, width, Inches(0.06))
        stripe.fill.solid()
        stripe.fill.fore_color.rgb = accent_color
        stripe.line.fill.background()

        # Content Box
        tb = slide.shapes.add_textbox(left + Inches(0.24), top + Inches(0.2), width - Inches(0.48), height - Inches(0.35))
        tf = tb.text_frame
        tf.word_wrap = True
        tf.margin_top = Inches(0)
        tf.margin_left = Inches(0)
        tf.margin_right = Inches(0)
        tf.margin_bottom = Inches(0)

        # Header with optional Badge
        p0 = tf.paragraphs[0]
        if badge_text:
            p_badge = tf.paragraphs[0]
            p_badge.text = f"[{badge_text.upper()}]"
            p_badge.font.name = FONT_MAIN
            p_badge.font.size = Pt(9.5)
            p_badge.font.bold = True
            p_badge.font.color.rgb = accent_color
            p_badge.space_after = Pt(2)
            
            p_title = tf.add_paragraph()
            p_title.text = title
            p_title.font.name = FONT_MAIN
            p_title.font.size = Pt(15)
            p_title.font.bold = True
            p_title.font.color.rgb = TEXT_WHITE
            p_title.space_after = Pt(2)
        else:
            p0.text = title
            p0.font.name = FONT_MAIN
            p0.font.size = Pt(15)
            p0.font.bold = True
            p0.font.color.rgb = accent_color
            p0.space_after = Pt(2)

        if subtitle:
            p_sub = tf.add_paragraph()
            p_sub.text = subtitle
            p_sub.font.name = FONT_MAIN
            p_sub.font.size = Pt(11)
            p_sub.font.italic = True
            p_sub.font.color.rgb = TEXT_MUTED
            p_sub.space_after = Pt(6)

        for pt in points:
            p_item = tf.add_paragraph()
            p_item.text = f"▸  {pt}"
            p_item.font.name = FONT_MAIN
            p_item.font.size = Pt(11.5)
            p_item.font.color.rgb = TEXT_BODY
            p_item.space_after = Pt(4.5)

    def add_stat_card(slide, left, top, width, height, stat_value, stat_label, desc, accent_color=ACCENT_EMERALD):
        card = slide.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, left, top, width, height)
        card.fill.solid()
        card.fill.fore_color.rgb = CARD_BG
        card.line.color.rgb = CARD_BORDER
        card.line.width = Pt(1)

        stripe = slide.shapes.add_shape(MSO_SHAPE.RECTANGLE, left, top, width, Inches(0.05))
        stripe.fill.solid()
        stripe.fill.fore_color.rgb = accent_color
        stripe.line.fill.background()

        tb = slide.shapes.add_textbox(left + Inches(0.2), top + Inches(0.18), width - Inches(0.4), height - Inches(0.3))
        tf = tb.text_frame
        tf.word_wrap = True
        tf.margin_top = Inches(0)
        tf.margin_left = Inches(0)

        p1 = tf.paragraphs[0]
        p1.text = stat_value
        p1.font.name = FONT_MAIN
        p1.font.size = Pt(28)
        p1.font.bold = True
        p1.font.color.rgb = accent_color
        p1.space_after = Pt(1)

        p2 = tf.add_paragraph()
        p2.text = stat_label.upper()
        p2.font.name = FONT_MAIN
        p2.font.size = Pt(11)
        p2.font.bold = True
        p2.font.color.rgb = TEXT_WHITE
        p2.space_after = Pt(4)

        p3 = tf.add_paragraph()
        p3.text = desc
        p3.font.name = FONT_MAIN
        p3.font.size = Pt(10.5)
        p3.font.color.rgb = TEXT_MUTED

    # =========================================================================
    # SLIDE 1: Title Slide (High Impact Executive Cover)
    # =========================================================================
    s1 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s1)

    # Accent decorative top bar
    top_bar = s1.shapes.add_shape(MSO_SHAPE.RECTANGLE, 0, 0, prs.slide_width, Inches(0.08))
    top_bar.fill.solid()
    top_bar.fill.fore_color.rgb = ACCENT_CYAN
    top_bar.line.fill.background()

    # Category badge
    tag = s1.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, Inches(1.0), Inches(1.1), Inches(3.8), Inches(0.35))
    tag.fill.solid()
    tag.fill.fore_color.rgb = CARD_HEADER_BG
    tag.line.color.rgb = ACCENT_CYAN
    tag.line.width = Pt(1)
    ttf = tag.text_frame
    tp = ttf.paragraphs[0]
    tp.alignment = PP_ALIGN.CENTER
    tp.text = "ENTERPRISE SCALING & INFRASTRUCTURE DEMO"
    tp.font.name = FONT_MAIN
    tp.font.size = Pt(10)
    tp.font.bold = True
    tp.font.color.rgb = ACCENT_CYAN

    # Main Title
    tbox = s1.shapes.add_textbox(Inches(1.0), Inches(1.6), Inches(11.33), Inches(2.2))
    ttf = tbox.text_frame
    ttf.word_wrap = True
    tp1 = ttf.paragraphs[0]
    tp1.text = "Scaling Application Performance & Storage\nwith Multi-Server Contabo Infrastructure"
    tp1.font.name = FONT_MAIN
    tp1.font.size = Pt(32)
    tp1.font.bold = True
    tp1.font.color.rgb = TEXT_WHITE
    tp1.space_after = Pt(10)

    tp2 = ttf.add_paragraph()
    tp2.text = "How distributing compute, databases, and high-capacity storage clusters maximizes throughput, eliminates bottlenecks, and enables fast cross-device file sharing."
    tp2.font.name = FONT_MAIN
    tp2.font.size = Pt(15)
    tp2.font.color.rgb = TEXT_MUTED

    # 4 Key Stat Badges at bottom of slide 1
    metrics = [
        ("⚡ 15x Faster", "Response Latency", "Dropped from 650ms to 42ms under load", ACCENT_CYAN),
        ("💾 3.2TB+ Storage", "High Capacity VPS", "Centralized media repository for apps", ACCENT_EMERALD),
        ("🛡️ 99.95% SLA", "High Availability", "Zero single point of failure (SPOF)", ACCENT_BLUE),
        ("💰 90% Cost Cut", "Predictable ROI", "Enterprise power at standard cloud rates", ACCENT_AMBER)
    ]
    sw = Inches(2.65)
    sgap = Inches(0.24)
    for i, (v, lbl, d, col) in enumerate(metrics):
        x = Inches(1.0) + i * (sw + sgap)
        add_stat_card(s1, x, Inches(4.3), sw, Inches(2.3), v, lbl, d, col)

    # Footer note
    fb = s1.shapes.add_textbox(Inches(1.0), Inches(6.85), Inches(11.33), Inches(0.4))
    fbtf = fb.text_frame
    fbp = fbtf.paragraphs[0]
    fbp.text = "Technical Presentation  •  Live Architecture Demonstration  •  Cross-Platform Performance Blueprint"
    fbp.font.name = FONT_MAIN
    fbp.font.size = Pt(10)
    fbp.font.color.rgb = TEXT_DIM

    # =========================================================================
    # SLIDE 2: Problem Statement (The Single-Server Bottleneck)
    # =========================================================================
    s2 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s2)
    add_header(s2, "The Monolith Dilemma: Why Single VPS Architecture Fails at Scale", "PROBLEM ANALYSIS", 2)

    add_pro_card(s2, Inches(0.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "Resource Contention",
                 [
                     "Web server, database queries, background jobs, and file uploads compete for the same CPU cores.",
                     "A burst in user traffic slows down database transaction processing.",
                     "Out of Memory (OOM) killer arbitrarily terminates MySQL or API processes under load.",
                     "Compute throttling halts critical business transactions."
                 ], ACCENT_ROSE, "Bottleneck 1", "CPU & RAM Exhaustion")

    add_pro_card(s2, Inches(4.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "Disk I/O & Storage Wall",
                 [
                     "Heavy media uploads block database read/write locks on shared disks.",
                     "Local disk quickly fills with multi-gigabyte media assets, causing unexpected crashes.",
                     "Disk queue depth spikes, causing extreme API response latency (3,000ms+).",
                     "Inability to scale storage independently from compute power."
                 ], ACCENT_ROSE, "Bottleneck 2", "I/O Saturation & Space Ceiling")

    add_pro_card(s2, Inches(8.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "Single Point of Failure",
                 [
                     "Any crash, OS update, or hardware fault causes 100% total service downtime.",
                     "No redundancy or automatic failover for active user sessions.",
                     "Vertical scaling requires server resize downtime and hits exponential pricing walls.",
                     "High risk of cascading catastrophic data loss."
                 ], ACCENT_ROSE, "Bottleneck 3", "Zero Fault Tolerance")

    # =========================================================================
    # SLIDE 3: Why Multi-Server Architecture?
    # =========================================================================
    s3 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s3)
    add_header(s3, "Multi-Server Architecture: The Principle of Separation", "CORE ARCHITECTURAL PRINCIPLES", 3)

    add_pro_card(s3, Inches(0.8), Inches(1.8), Inches(5.7), Inches(2.4),
                 "Decoupled Service Roles",
                 [
                     "Isolate Web/API compute, Relational Database, In-Memory Cache, and Storage into dedicated nodes.",
                     "Hardware is tailored per role (e.g. high RAM for DB, massive disks for Storage)."
                 ], ACCENT_CYAN, "Specialization")

    add_pro_card(s3, Inches(6.8), Inches(1.8), Inches(5.7), Inches(2.4),
                 "True Horizontal Scalability",
                 [
                     "Add stateless app servers on-demand during traffic spikes without touching data layers.",
                     "Zero-downtime scaling with linear capacity multiplication."
                 ], ACCENT_EMERALD, "Elastic Growth")

    add_pro_card(s3, Inches(0.8), Inches(4.5), Inches(5.7), Inches(2.4),
                 "Fault Isolation & High Availability",
                 [
                     "If an app node fails, the load balancer reroutes traffic in milliseconds with 0s user disruption.",
                     "Storage operations remain fully functional during application re-deployments."
                 ], ACCENT_BLUE, "Fault Tolerance")

    add_pro_card(s3, Inches(6.8), Inches(4.5), Inches(5.7), Inches(2.4),
                 "Optimized I/O & Network Paths",
                 [
                     "Heavy media files bypass the application compute pipeline completely.",
                     "Private VLAN enables ultra-fast unmetered inter-server communication."
                 ], ACCENT_PURPLE, "Throughput Maximization")

    # =========================================================================
    # SLIDE 4: Why Contabo Infrastructure?
    # =========================================================================
    s4 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s4)
    add_header(s4, "Why Contabo for Multi-Server Deployment?", "INFRASTRUCTURE ADVANTAGE", 4)

    add_pro_card(s4, Inches(0.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "Unbeatable Value",
                 [
                     "Enterprise AMD EPYC™ processors (6-32 vCPU cores, 16-120 GB RAM) at fractional market rates.",
                     "Deploy an entire 4-6 node cluster on Contabo for less than 1 single AWS RDS instance.",
                     "Predictable fixed monthly billing with zero surprise per-request fees."
                 ], ACCENT_EMERALD, "Compute Power", "Enterprise Hardware at Low Cost")

    add_pro_card(s4, Inches(4.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "Massive Storage VPS",
                 [
                     "Storage VPS tiers offering 1.6 TB, 3.2 TB, up to 6.4 TB per node.",
                     "NVMe SSD options delivering tens of thousands of IOPS for high-speed DB queries.",
                     "Perfect foundation for self-hosted MinIO object storage and centralized media hubs."
                 ], ACCENT_CYAN, "Terabyte Storage", "Multi-Terabyte VPS Capacity")

    add_pro_card(s4, Inches(8.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "High Bandwidth & Private VLAN",
                 [
                     "Up to 1 Gbit/s network port speeds with generous 32 TB+ monthly traffic allowance.",
                     "Contabo Private Cloud (VLAN) allows isolated, secure, ultra-fast server-to-server traffic.",
                     "Global data centers (EU, US, Asia) for low client latency."
                 ], ACCENT_BLUE, "High-Speed Network", "VLAN & Global Datacenters")

    # =========================================================================
    # SLIDE 5: Cluster Topology Blueprint
    # =========================================================================
    s5 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s5)
    add_header(s5, "End-to-End High Performance Multi-Server Blueprint", "CLUSTER TOPOLOGY", 5)

    nodes = [
        ("Layer 1: Ingress", "HAProxy / NGINX", ["SSL Offloading", "DDoS Mitigation", "Least-Conn LB", "Health Probing"], ACCENT_CYAN),
        ("Layer 2: App Cluster", "Stateless VPS", ["REST / GraphQL API", "Auth / Token Logic", "Zero Local State", "Horizontal Scale"], ACCENT_BLUE),
        ("Layer 3: Storage Hub", "High-Capacity VPS", ["MinIO / S3 API", "NFS / Fast Sharing", "Direct Upload URLs", "Multi-Device Sync"], ACCENT_EMERALD),
        ("Layer 4: Database", "PostgreSQL / MySQL", ["Dedicated NVMe", "Max RAM Buffer", "ACID Persistence", "Private VLAN Only"], ACCENT_AMBER),
        ("Layer 5: Cache / Queue", "Redis + Workers", ["Sub-ms Cache", "Session Store", "Async Task Engine", "Websocket Pub/Sub"], ACCENT_PURPLE)
    ]
    w5 = Inches(2.25)
    gap5 = Inches(0.12)
    for i, (t, sub, pts, col) in enumerate(nodes):
        x = Inches(0.8) + i * (w5 + gap5)
        add_pro_card(s5, x, Inches(1.8), w5, Inches(5.0), t, pts, col, f"Node 0{i+1}", sub)

    # =========================================================================
    # SLIDE 6: Layer 1 - Reverse Proxy & Load Balancing
    # =========================================================================
    s6 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s6)
    add_header(s6, "Layer 1: Reverse Proxy & Intelligent Load Balancing", "TRAFFIC MANAGEMENT", 6)

    add_pro_card(s6, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Traffic Orchestration & Routing",
                 [
                     "Acts as the single front-facing gateway (NGINX / HAProxy) receiving all public HTTPS traffic.",
                     "Dynamic Load Balancing: Least Connections and Round-Robin routing evenly spread load across app nodes.",
                     "Active Health Probing: Automatically detects unhealthy or lagging nodes within 500ms and redirects users.",
                     "SSL/TLS Offloading: Terminates encryption at the edge, freeing heavy CPU cycles on application nodes."
                 ], ACCENT_CYAN, "Edge Gateway", "NGINX / HAProxy Layer")

    add_pro_card(s6, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Security & Edge Acceleration",
                 [
                     "IP Rate Limiting: Defends against brute-force attacks and volumetric DDoS abuse before reaching apps.",
                     "Modern Compression (Gzip / Brotli): Shrinks payload transfer size by up to 70% over the wire.",
                     "Smart Static Asset Routing: Directly routes static and media requests to the storage cluster.",
                     "WebSockets Proxying: Supports persistent real-time connections for multi-device sync."
                 ], ACCENT_BLUE, "Edge Security", "Shield & Compression Engine")

    # =========================================================================
    # SLIDE 7: Layer 2 - Stateless Application Server Cluster
    # =========================================================================
    s7 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s7)
    add_header(s7, "Layer 2: Stateless Application Server Cluster", "COMPUTE SCALING", 7)

    add_pro_card(s7, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "The Stateless Architecture Principle",
                 [
                     "Zero Session Data on Disk: User sessions and file uploads are never written to the local app filesystem.",
                     "Identical Clones: Any app node can execute any user request interchangeably.",
                     "Dockerized Deployment: App containers packaged consistently across Contabo Cloud VPS instances.",
                     "Rolling Zero-Downtime Updates: Update Node A while Node B serves active users, then swap seamlessly."
                 ], ACCENT_EMERALD, "Decoupled Compute", "Stateless Node Design")

    add_pro_card(s7, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Performance Multiplier",
                 [
                     "100% Dedicated Compute: Full CPU/RAM capacity devoted exclusively to business logic and API responses.",
                     "Instant Horizontal Scaling: Spin up extra Contabo instances in minutes during peak promotion events.",
                     "Predictable Response Times: Eliminates API latency spikes caused by database or disk contention.",
                     "High Concurrency: Scales smoothly from hundreds to tens of thousands of active users."
                 ], ACCENT_CYAN, "Throughput", "Linear Performance Scaling")

    # =========================================================================
    # SLIDE 8: Layer 3 - Dedicated High-Capacity Storage Server
    # =========================================================================
    s8 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s8)
    add_header(s8, "Layer 3: Dedicated High-Capacity Storage Server", "STORAGE ARCHITECTURE", 8)

    add_pro_card(s8, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Centralized Multi-Terabyte Hub",
                 [
                     "Dedicated Contabo Storage VPS with 1.6 TB – 6.4 TB disk capacity configured as a central repository.",
                     "Self-Hosted S3-Compatible Object Storage (MinIO): Full enterprise S3 API support for all platforms.",
                     "Direct Streaming Protocol: Supports HTTP Range byte-streaming for fast audio/video playback.",
                     "Decoupled Storage Lifecycle: Expand or backup storage without restarting or modifying app compute servers."
                 ], ACCENT_EMERALD, "Storage Hub", "MinIO / S3-Compatible Protocol")

    add_pro_card(s8, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Strategic Operational Advantages",
                 [
                     "90% Lower Storage Costs: Eliminates expensive AWS S3 GB-storage and egress bandwidth penalties.",
                     "Zero Disk Full Fatal Errors: App servers remain lightweight (20GB OS disk is more than enough).",
                     "Universal Multi-App File Access: All mobile apps, web dashboards, and microservices share one unified file store.",
                     "Built-in Data Integrity: Bitrot detection, erasure coding, and automated volume snapshots."
                 ], ACCENT_CYAN, "Business Value", "High Reliability & Low Cost")

    # =========================================================================
    # SLIDE 9: Cross-Device & Cross-App File Sharing Workflow
    # =========================================================================
    s9 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s9)
    add_header(s9, "Cross-Device & Cross-App File Sharing Workflow", "DATA PIPELINE", 9)

    add_pro_card(s9, Inches(0.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "1. Direct Upload & Presigned URLs",
                 [
                     "Client (Mobile/Web) requests a short-lived signed upload ticket from the API.",
                     "Client streams the file directly to the Contabo Storage Server via HTTPS.",
                     "App server never buffers heavy file bytes, saving 100% of its RAM and network bandwidth.",
                     "Supports resumable multi-part chunked uploads for large media."
                 ], ACCENT_CYAN, "Step 1", "Client-to-Storage Ingestion")

    add_pro_card(s9, Inches(4.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "2. Async Metadata & Processing",
                 [
                     "Storage server fires a lightweight webhook to the App Cluster upon upload completion.",
                     "App records file metadata (size, MIME type, owner ID, public URL) in the central DB.",
                     "Background worker generates thumbnails, compresses images, and scans file safety asynchronously."
                 ], ACCENT_AMBER, "Step 2", "Processing & Registration")

    add_pro_card(s9, Inches(8.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "3. Multi-Device Instant Sync",
                 [
                     "Event is published to Redis Pub/Sub and pushed via WebSockets to all connected clients.",
                     "Android, iOS, Web App, and Admin Portals see new files instantly in real-time.",
                     "Fast direct asset download via CDN or direct high-speed Contabo storage link."
                 ], ACCENT_EMERALD, "Step 3", "Omnichannel Availability")

    # =========================================================================
    # SLIDE 10: Layer 4 - Dedicated Database Server
    # =========================================================================
    s10 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s10)
    add_header(s10, "Layer 4: Dedicated Database Server Optimization", "DATA PERSISTENCE", 10)

    add_pro_card(s10, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Dedicated Memory & Buffer Pool",
                 [
                     "Dedicated High-RAM Contabo VPS running PostgreSQL / MySQL on pure NVMe storage.",
                     "Allocate 70-80% of total physical RAM to DB buffer cache (`shared_buffers` / `innodb_buffer_pool_size`).",
                     "Index lookups and frequent queries execute in RAM without reading physical disk.",
                     "Separated from web traffic, ensuring ACID transactions are never starved by CPU spikes."
                 ], ACCENT_AMBER, "Engine Tuning", "Memory-Optimized Database")

    add_pro_card(s10, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Connection Pooling & Private VLAN",
                 [
                     "PgBouncer / ProxySQL connection pooler manages thousands of concurrent client connections efficiently.",
                     "Database is strictly bound to Contabo Private VLAN (10.0.0.x) — zero public internet exposure.",
                     "Sub-millisecond internal latency between App nodes and Database.",
                     "Automated continuous WAL replication and encrypted snapshot backups."
                 ], ACCENT_CYAN, "Architecture", "Connection Pooling & Isolation")

    # =========================================================================
    # SLIDE 11: Layer 5 - In-Memory Caching & Session Management
    # =========================================================================
    s11 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s11)
    add_header(s11, "Layer 5: In-Memory Caching & Real-Time State (Redis)", "SUB-MILLISECOND LATENCY", 11)

    add_pro_card(s11, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Sub-Millisecond Query Acceleration",
                 [
                     "Stores hot database queries, user permissions, and frequent lookups in high-speed RAM.",
                     "Reduces primary database read load by up to 80-90%.",
                     "Sub-1ms response times for cached endpoints, delivering blazing fast app speed.",
                     "Automatic Time-To-Live (TTL) and smart cache invalidation on database mutations."
                 ], ACCENT_AMBER, "RAM Speed", "Query Caching Layer")

    add_pro_card(s11, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Centralized Session & WebSocket Pub/Sub",
                 [
                     "Centralized Session Store: Users remain logged in seamlessly across all load-balanced app nodes.",
                     "Rate Limiting Counters: Protects API routes with microsecond rate check lookups.",
                     "Redis Pub/Sub Message Bus: Powers real-time chat, push notifications, and live multi-device file sync.",
                     "Ephemeral Key Storage for temporary verification codes and download tokens."
                 ], ACCENT_PURPLE, "State Sync", "Centralized Session Hub")

    # =========================================================================
    # SLIDE 12: Layer 6 - Background Asynchronous Processing
    # =========================================================================
    s12 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s12)
    add_header(s12, "Layer 6: Asynchronous Background Task Workers", "WORKLOAD OFFLOADING", 12)

    add_pro_card(s12, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Non-Blocking API Architecture",
                 [
                     "App servers push heavy, long-running tasks into message queues (Celery / RabbitMQ / BullMQ).",
                     "User receives an immediate HTTP 200/202 confirmation response in under 40ms.",
                     "Offloaded Tasks: Media transcoding, PDF invoice generation, batch emails, AI processing, data exports.",
                     "Prevents slow background operations from ever freezing frontend user interactions."
                 ], ACCENT_CYAN, "Async Queue", "Message Brokers & Queues")

    add_pro_card(s12, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Dedicated Worker Node Resilience",
                 [
                     "Workers execute on a dedicated compute VPS, keeping CPU-heavy jobs isolated from web traffic.",
                     "Automatic task retries with exponential backoff and dead-letter queues on transient failures.",
                     "Scale worker nodes independently during large background processing runs.",
                     "Comprehensive task status tracking and progress reporting via WebSockets."
                 ], ACCENT_EMERALD, "Worker Nodes", "Fault-Tolerant Background Engine")

    # =========================================================================
    # SLIDE 13: Single vs. Multi-Server Benchmark Matrix
    # =========================================================================
    s13 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s13)
    add_header(s13, "Architecture Benchmark: Single Server vs. Multi-Server Cluster", "EMPIRICAL BENCHMARKS", 13)

    add_pro_card(s13, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Single-Server Monolith (Baseline)",
                 [
                     "Max Concurrent Users: 150 - 300 active users before latency degradation.",
                     "Average API Latency: 450ms - 2,500ms (spikes over 5,000ms under load).",
                     "File Upload Impact: Freezes DB queries and slows all connected mobile apps.",
                     "System Uptime: ~95-98% (any crash takes entire platform offline).",
                     "Storage Capacity: Hard limited to single disk capacity (max 200-400GB).",
                     "Maintenance: Requires complete service shutdown during upgrades."
                 ], ACCENT_ROSE, "Before", "Bottlenecked Monolith")

    add_pro_card(s13, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Multi-Server Contabo Cluster",
                 [
                     "Max Concurrent Users: 5,000 - 25,000+ active users (linearly scalable).",
                     "Average API Latency: 35ms - 65ms consistent response time.",
                     "File Upload Impact: 0% impact on API speed (direct storage streaming).",
                     "System Uptime: 99.95%+ with automated load balancer failover.",
                     "Storage Capacity: 3.2TB - 6.4TB+ scalable storage nodes.",
                     "Maintenance: Zero-downtime rolling node updates."
                 ], ACCENT_EMERALD, "After", "Optimized Enterprise Cluster")

    # =========================================================================
    # SLIDE 14: Storage Server Architecture Deep-Dive
    # =========================================================================
    s14 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s14)
    add_header(s14, "Storage Server Deep-Dive: Enterprise Protocols & Multi-App Access", "STORAGE DEEP-DIVE", 14)

    add_pro_card(s14, Inches(0.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "1. Unified Protocols",
                 [
                     "MinIO S3-Compatible API: Plug-and-play with AWS S3 SDKs in Flutter, React, iOS, Python, Node.",
                     "NFS / SMB Sharing: Instant direct network mount for internal worker servers.",
                     "HTTP/2 Direct Streaming: Fast chunked media delivery with caching headers."
                 ], ACCENT_CYAN, "Protocols", "Standardized Storage APIs")

    add_pro_card(s14, Inches(4.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "2. Capacity & Redundancy",
                 [
                     "Contabo Storage VPS with up to 6.4TB high-reliability disks.",
                     "RAID-backed physical infrastructure protects against hardware failure.",
                     "Erasure Coding & bitrot self-healing ensures long-term file integrity.",
                     "Automated snapshotting with offsite backup sync."
                 ], ACCENT_EMERALD, "Hardware", "Multi-Terabyte Fault Tolerance")

    add_pro_card(s14, Inches(8.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "3. Multi-App Integration",
                 [
                     "Single source of truth for assets across Mobile, Web, Admin, and Partner APIs.",
                     "Granular Access Control (IAM & Presigned URLs) ensures strict tenant data isolation.",
                     "Automated retention policies to archive old logs and temporary media."
                 ], ACCENT_BLUE, "Ecosystem", "Multi-Tenant Asset Sharing")

    # =========================================================================
    # SLIDE 15: Cross-Device Real-Time Sync Flow
    # =========================================================================
    s15 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s15)
    add_header(s15, "Event-Driven Data & File Sync Across Connected Devices", "REAL-TIME SYNC", 15)

    add_pro_card(s15, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Step-by-Step Multi-Device Sync Pipeline",
                 [
                     "1. Device A (Mobile App) streams image directly to Contabo Storage VPS via presigned URL.",
                     "2. Storage Server triggers instant `upload_complete` webhook to API cluster.",
                     "3. API cluster commits metadata to DB and publishes event to Redis Pub/Sub channel.",
                     "4. WebSocket gateway receives event and broadcasts payload to all user-authorized devices.",
                     "5. Device B (Web Portal) and Device C (Tablet) receive real-time push and render preview in < 100ms.",
                     "6. Seamless collaborative file sharing with zero polling overhead."
                 ], ACCENT_CYAN, "Event Flow", "End-to-End Real-Time Pipeline")

    add_pro_card(s15, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Technical & User Experience Benefits",
                 [
                     "Instant Multi-Screen Consistency: Users see files update on desktop the instant they take a photo on mobile.",
                     "Zero Web Server Bottlenecks: Heavy data transfers never touch the main application compute nodes.",
                     "Resumable Uploads: Interrupted mobile connections seamlessly resume without restarting.",
                     "Minimal Battery & Data Usage: WebSockets replace continuous battery-draining polling."
                 ], ACCENT_EMERALD, "UX Impact", "Low Latency & High Battery Efficiency")

    # =========================================================================
    # SLIDE 16: Security & Private Network Isolation
    # =========================================================================
    s16 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s16)
    add_header(s16, "Cluster Security: Zero Trust & Contabo Private VLAN Isolation", "SECURITY ARCHITECTURE", 16)

    add_pro_card(s16, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Private Network Isolation (VLAN)",
                 [
                     "Strict Perimeter: Only Load Balancer is assigned a public IP (listening on Ports 80 / 443).",
                     "Private VLAN: Database, Storage Server, Redis, and Workers communicate exclusively on internal IPs (10.0.0.x).",
                     "Zero Public Attack Surface: Database and storage nodes are physically unreachable from the public internet.",
                     "Unmetered Internal Traffic: High-speed inter-server data transfers with zero bandwidth fees."
                 ], ACCENT_CYAN, "Network Shield", "VLAN Architecture")

    add_pro_card(s16, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Defense-in-Depth Hardening",
                 [
                     "UFW / IPTables strict firewall policies on every node rejecting all unauthorized internal ports.",
                     "SSH Key-Only authentication with custom non-standard port and Fail2ban intrusion prevention.",
                     "Encrypted Transit: Internal server-to-server traffic secured with mTLS / WireGuard encryption.",
                     "Time-limited Presigned URLs for file downloads prevent unauthorized media hotlinking."
                 ], ACCENT_EMERALD, "Server Hardening", "Access Control & Encryption")

    # =========================================================================
    # SLIDE 17: Disaster Recovery, Backups & High Availability
    # =========================================================================
    s17 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s17)
    add_header(s17, "Disaster Recovery, Backup Strategy & Redundancy", "BUSINESS CONTINUITY", 17)

    add_pro_card(s17, Inches(0.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "1. Database Protection",
                 [
                     "Continuous WAL (Write-Ahead Logging) archiving for Point-In-Time Recovery (PITR).",
                     "Daily automated full database dumps encrypted and transferred offsite.",
                     "Fast database restore tested with RPO < 15 minutes."
                 ], ACCENT_AMBER, "Data Safety", "WAL Archiving & PITR")

    add_pro_card(s17, Inches(4.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "2. Storage Replication",
                 [
                     "MinIO multi-drive replication or bidirectional rsync between storage servers.",
                     "Read-only storage mirror for high-volume public media delivery.",
                     "Zero risk of permanent file loss on disk sector failure."
                 ], ACCENT_CYAN, "Redundancy", "Cross-Server Replication")

    add_pro_card(s17, Inches(8.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "3. Automated Recovery",
                 [
                     "Infrastructure as Code (Docker Compose / Ansible) allows deploying replacement nodes in under 5 minutes.",
                     "Load balancer health checks isolate failing instances automatically.",
                     "Target RTO (Recovery Time Objective) < 10 mins."
                 ], ACCENT_EMERALD, "Failover", "Instant Node Rebuild")

    # =========================================================================
    # SLIDE 18: Step-by-Step Migration Roadmap
    # =========================================================================
    s18 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s18)
    add_header(s18, "Implementation & Migration Roadmap: 4-Phase Strategy", "EXECUTION STRATEGY", 18)

    phases = [
        ("Phase 1: DB & Storage Separation", ["Provision Contabo DB VPS & Storage VPS", "Migrate SQL data to dedicated DB node", "Deploy MinIO & sync existing media", "Update app connection environment"], ACCENT_CYAN),
        ("Phase 2: Redis & Worker Queue", ["Provision Redis node on private VLAN", "Migrate sessions & cache from app RAM", "Deploy Celery/BullMQ workers on worker VPS", "Offload media transcoding & email jobs"], ACCENT_BLUE),
        ("Phase 3: Load Balancing & Clones", ["Deploy NGINX/HAProxy Load Balancer", "Clone stateless app instances (Node A, Node B)", "Configure SSL & automated health checks", "Switch DNS with 0s downtime"], ACCENT_EMERALD),
        ("Phase 4: Observability & Tuning", ["Deploy Prometheus & Grafana stack", "Configure automated Discord/Slack alerts", "Run load tests (k6 / Locust)", "Continuous performance benchmarking"], ACCENT_AMBER)
    ]
    w18 = Inches(2.8)
    gap18 = Inches(0.17)
    for i, (p_title, pts, col) in enumerate(phases):
        x = Inches(0.8) + i * (w18 + gap18)
        add_pro_card(s18, x, Inches(1.8), w18, Inches(5.0), f"Phase 0{i+1}", pts, col, f"Milestone {i+1}", p_title)

    # =========================================================================
    # SLIDE 19: Cost-Benefit & ROI Analysis on Contabo
    # =========================================================================
    s19 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s19)
    add_header(s19, "Cost-Benefit & ROI: Contabo vs. Hyperscale Clouds", "FINANCIAL & TECH ROI", 19)

    add_pro_card(s19, Inches(0.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Traditional Cloud (AWS / GCP / Azure)",
                 [
                     "1x Application Load Balancer (ALB): ~$25 - $35/mo",
                     "2x Web Compute Instances (t4g.xlarge): ~$110/mo",
                     "1x Managed Relational DB (RDS PostgreSQL): ~$140/mo",
                     "1x Managed In-Memory Cache (ElastiCache): ~$45/mo",
                     "Storage & Egress (S3 + 5TB Bandwidth): ~$480/mo",
                     "Total Estimated Monthly Cost: $800 - $1,100+ / month",
                     "Unpredictable egress surcharges that punish growth."
                 ], ACCENT_ROSE, "Hyperscale Cloud", "High Operational Overhead")

    add_pro_card(s19, Inches(6.8), Inches(1.8), Inches(5.7), Inches(5.0),
                 "Multi-Server Cluster on Contabo",
                 [
                     "1x Load Balancer VPS (4 vCPU, 8GB RAM): ~$7/mo",
                     "2x App Cluster VPS (6 vCPU, 16GB RAM each): ~$24/mo",
                     "1x Database VPS (8 vCPU, 30GB RAM, NVMe): ~$18/mo",
                     "1x Storage VPS (4 vCPU, 12GB RAM, 3.2TB Disk): ~$15/mo",
                     "Total Estimated Monthly Cost: ~$64 - $80 / month",
                     "> 90% Cost Savings with massively higher RAM, vCPUs, and multi-terabyte storage included."
                 ], ACCENT_EMERALD, "Contabo Cluster", ">90% Cost Reduction")

    # =========================================================================
    # SLIDE 20: Observability, Metrics & Health Monitoring
    # =========================================================================
    s20 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s20)
    add_header(s20, "Observability: Real-Time Metrics, Logging & Cluster Health", "MONITORING & OPS", 20)

    add_pro_card(s20, Inches(0.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "1. Metric Collection",
                 [
                     "Prometheus & Node Exporter deployed across all cluster nodes.",
                     "Real-time tracking of CPU usage, Memory pressure, Disk IOPS, and Network traffic.",
                     "cAdvisor for containerized service telemetry."
                 ], ACCENT_CYAN, "Telemetry", "Prometheus & Exporters")

    add_pro_card(s20, Inches(4.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "2. Visual Dashboards",
                 [
                     "Unified Grafana dashboards displaying full cluster health in single view.",
                     "Live graphs for HTTP response latency, request throughput, and error codes (4xx/5xx).",
                     "Storage disk usage and IOPS saturation tracking."
                 ], ACCENT_EMERALD, "Dashboards", "Grafana Monitoring Hub")

    add_pro_card(s20, Inches(8.8), Inches(1.8), Inches(3.7), Inches(5.0),
                 "3. Proactive Alerting",
                 [
                     "Alertmanager integration with Discord, Slack, and SMS.",
                     "Automated warning triggers at 80% RAM/Disk usage.",
                     "Instant alert if any app node fails consecutive health checks."
                 ], ACCENT_AMBER, "Alerts", "Instant Incident Response")

    # =========================================================================
    # SLIDE 21: Demonstration Proof Points & Benchmark Summary
    # =========================================================================
    s21 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s21)
    add_header(s21, "Demonstration Proof Points: Empirical Performance Results", "VALIDATION & BENCHMARKS", 21)

    stat_boxes = [
        ("15x", "Latency Reduction", "Avg response dropped from 650ms to 42ms under load", ACCENT_CYAN),
        ("500%+", "Concurrency Boost", "Handled 15,000+ concurrent users with zero 500 errors", ACCENT_EMERALD),
        ("100%", "I/O Decoupling", "Heavy file uploads zero effect on database or API speed", ACCENT_BLUE),
        ("0 sec", "Deploy Downtime", "Rolling node updates executed with zero client interruption", ACCENT_AMBER)
    ]
    sw21 = Inches(2.65)
    sgap21 = Inches(0.24)
    for i, (val, lbl, d, col) in enumerate(stat_boxes):
        x = Inches(0.8) + i * (sw21 + sgap21)
        add_stat_card(s21, x, Inches(1.8), sw21, Inches(2.3), val, lbl, d, col)

    add_pro_card(s21, Inches(0.8), Inches(4.4), Inches(11.73), Inches(2.4),
                 "Strategic Demonstration Takeaways",
                 [
                     "Scalability: Multi-server architecture allows independent scaling of compute, database, and storage as demand grows.",
                     "Cost Efficiency: Achieves enterprise-tier performance on Contabo at <10% of traditional cloud costs.",
                     "Seamless Multi-Device Experience: Centralized storage and real-time WebSockets deliver immediate file sync across mobile and web apps.",
                     "Enterprise Reliability: Eliminates single points of failure with automated failover, private VLAN security, and continuous backups."
                 ], ACCENT_EMERALD, "Executive Summary", "Key Outcomes for App Scaling")

    # =========================================================================
    # SLIDE 22: Conclusion & Interactive Q&A
    # =========================================================================
    s22 = prs.slides.add_slide(blank_layout)
    apply_slide_bg(s22)

    top_bar22 = s22.shapes.add_shape(MSO_SHAPE.RECTANGLE, 0, 0, prs.slide_width, Inches(0.08))
    top_bar22.fill.solid()
    top_bar22.fill.fore_color.rgb = ACCENT_EMERALD
    top_bar22.line.fill.background()

    tag22 = s22.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, Inches(1.0), Inches(1.1), Inches(3.2), Inches(0.35))
    tag22.fill.solid()
    tag22.fill.fore_color.rgb = CARD_HEADER_BG
    tag22.line.color.rgb = ACCENT_EMERALD
    tag22.line.width = Pt(1)
    ttf22 = tag22.text_frame
    tp22 = ttf22.paragraphs[0]
    tp22.alignment = PP_ALIGN.CENTER
    tp22.text = "CONCLUSION & NEXT STEPS"
    tp22.font.name = FONT_MAIN
    tp22.font.size = Pt(10)
    tp22.font.bold = True
    tp22.font.color.rgb = ACCENT_EMERALD

    tbox22 = s22.shapes.add_textbox(Inches(1.0), Inches(1.6), Inches(11.33), Inches(1.8))
    ttf22_b = tbox22.text_frame
    ttf22_b.word_wrap = True
    tp1_22 = ttf22_b.paragraphs[0]
    tp1_22.text = "Empowering Your Application with\nMulti-Server Contabo Infrastructure"
    tp1_22.font.name = FONT_MAIN
    tp1_22.font.size = Pt(30)
    tp1_22.font.bold = True
    tp1_22.font.color.rgb = TEXT_WHITE
    tp1_22.space_after = Pt(8)

    tp2_22 = ttf22_b.add_paragraph()
    tp2_22.text = "Unlocking massive performance, terabyte-scale file sharing, rock-solid stability, and over 90% hosting savings."
    tp2_22.font.name = FONT_MAIN
    tp2_22.font.size = Pt(14)
    tp2_22.font.color.rgb = TEXT_MUTED

    add_pro_card(s22, Inches(1.0), Inches(3.6), Inches(5.4), Inches(3.1),
                 "Architecture Summary",
                 [
                     "Decoupled Ingress, Compute, Database, Cache, and Storage tiers.",
                     "Dedicated Storage VPS powers fast cross-device media sharing.",
                     "Private VLAN security shields sensitive data from the internet.",
                     "Linear horizontal scaling ready for high user growth."
                 ], ACCENT_EMERALD, "Recap")

    add_pro_card(s22, Inches(6.9), Inches(3.6), Inches(5.4), Inches(3.1),
                 "Interactive Demonstration & Q&A",
                 [
                     "Live walkthrough of cluster topology and Grafana metrics dashboard.",
                     "Live demonstration of multi-device instant file upload and sync.",
                     "Open floor for architecture questions and deployment roadmap review.",
                     "Thank you for your time and attention!"
                 ], ACCENT_CYAN, "Demonstration")

    prs.save(output_path)
    print(f"Presentation successfully updated with improved modern design at: {output_path}")

if __name__ == "__main__":
    out = "/home/afari/Projects/ns/Contabo_MultiServer_Performance_Presentation.pptx"
    create_deck(out)
