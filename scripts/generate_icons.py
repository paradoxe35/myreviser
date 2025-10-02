from PIL import Image, ImageDraw
import math
import os


def create_rounded_rectangle(
    draw, coords, radius, fill, shadow_offset=0, shadow_color=None
):
    """Draw a rounded rectangle with optional shadow"""
    x1, y1, x2, y2 = coords

    # Ensure coordinates are properly ordered
    x1, x2 = min(x1, x2), max(x1, x2)
    y1, y2 = min(y1, y2), max(y1, y2)

    # Ensure radius isn't too large
    radius = min(radius, (x2 - x1) // 2, (y2 - y1) // 2)

    # Draw shadow first if specified
    if shadow_offset and shadow_color:
        shadow_coords = (
            x1 + shadow_offset,
            y1 + shadow_offset,
            x2 + shadow_offset,
            y2 + shadow_offset,
        )
        # Shadow rectangle
        draw.rectangle(
            (
                shadow_coords[0] + radius,
                shadow_coords[1],
                shadow_coords[2] - radius,
                shadow_coords[3],
            ),
            fill=shadow_color,
        )
        draw.rectangle(
            (
                shadow_coords[0],
                shadow_coords[1] + radius,
                shadow_coords[2],
                shadow_coords[3] - radius,
            ),
            fill=shadow_color,
        )
        # Shadow corners
        for x, y in [
            (shadow_coords[0] + radius, shadow_coords[1] + radius),
            (shadow_coords[2] - radius, shadow_coords[1] + radius),
            (shadow_coords[0] + radius, shadow_coords[3] - radius),
            (shadow_coords[2] - radius, shadow_coords[3] - radius),
        ]:
            draw.ellipse(
                (x - radius, y - radius, x + radius, y + radius), fill=shadow_color
            )

    # Main rectangle
    draw.rectangle((x1 + radius, y1, x2 - radius, y2), fill=fill)
    draw.rectangle((x1, y1 + radius, x2, y2 - radius), fill=fill)

    # Corners
    for x, y in [
        (x1 + radius, y1 + radius),
        (x2 - radius, y1 + radius),
        (x1 + radius, y2 - radius),
        (x2 - radius, y2 - radius),
    ]:
        draw.ellipse((x - radius, y - radius, x + radius, y + radius), fill=fill)


def create_arrow(draw, start_pos, end_pos, width, fill):
    """Draw an arrow"""
    # Draw arrow body
    draw.line((start_pos, end_pos), fill=fill, width=width)

    # Calculate arrow head
    angle = math.atan2(end_pos[1] - start_pos[1], end_pos[0] - start_pos[0])
    arrow_length = width * 2
    arrow_angle = math.pi / 6  # 30 degrees

    x1 = end_pos[0] - arrow_length * math.cos(angle - arrow_angle)
    y1 = end_pos[1] - arrow_length * math.sin(angle - arrow_angle)
    x2 = end_pos[0] - arrow_length * math.cos(angle + arrow_angle)
    y2 = end_pos[1] - arrow_length * math.sin(angle + arrow_angle)

    # Ensure coordinates are properly ordered for polygon
    arrow_points = [
        (int(end_pos[0]), int(end_pos[1])),
        (int(x1), int(y1)),
        (int(x2), int(y2)),
    ]
    draw.polygon(arrow_points, fill=fill)


def create_image(
    width=256, height=256, primary_color="#4A90E2", accent_color="#2C3E50"
):
    """Create the Input Reviser icon with transparency"""
    # Create base image with transparency
    image = Image.new("RGBA", (width, height), (0, 0, 0, 0))
    draw = ImageDraw.Draw(image)

    # Calculate dimensions - no outer padding, fill entire canvas
    doc_height = height

    # Add slight shadow for depth
    shadow_color = (0, 0, 0, 60)  # Semi-transparent black
    shadow_offset = max(width // 40, 1)  # Ensure minimum shadow offset of 1

    # Calculate document offset for layered effect
    back_doc_offset = max(width // 20, 1)  # Ensure minimum offset
    front_doc_offset = max(width // 25, 1)  # Ensure minimum offset

    # Draw main document shapes
    # Back document with shadow (slightly offset)
    create_rounded_rectangle(
        draw,
        (
            back_doc_offset,
            back_doc_offset,
            width,
            height,
        ),
        max(width // 20, 1),  # Ensure minimum radius of 1
        accent_color,
        shadow_offset,
        shadow_color,
    )

    # Front document with shadow (fills most of canvas)
    create_rounded_rectangle(
        draw,
        (
            0,
            0,
            width - front_doc_offset,
            height - front_doc_offset,
        ),
        max(width // 20, 1),  # Ensure minimum radius of 1
        primary_color,
        shadow_offset,
        shadow_color,
    )

    # Draw "revision" arrows
    arrow_spacing = max(height // 6, 3)  # Ensure minimum spacing
    arrow_width = max(width // 40, 1)  # Ensure minimum width of 1
    arrow_color = "#FFFFFF"  # White arrows

    # Calculate arrow positions
    doc_center_x = (width - front_doc_offset) // 2
    doc_width_quarter = (width - front_doc_offset) // 4

    # Draw three arrows suggesting revision/transformation
    for i in range(3):
        y_pos = doc_height // 4 + (i * arrow_spacing)
        if i % 2 == 0:
            # Left to right arrow
            create_arrow(
                draw,
                (doc_center_x - doc_width_quarter, y_pos),
                (doc_center_x + doc_width_quarter, y_pos),
                arrow_width,
                arrow_color,
            )
        else:
            # Right to left arrow
            create_arrow(
                draw,
                (doc_center_x + doc_width_quarter, y_pos),
                (doc_center_x - doc_width_quarter, y_pos),
                arrow_width,
                arrow_color,
            )

    return image


# Example usage:
if __name__ == "__main__":
    # Get script directory
    script_dir = os.path.dirname(os.path.abspath(__file__))
    project_root = os.path.dirname(script_dir)
    assets_dir = os.path.join(project_root, "assets")
    build_dir = os.path.join(project_root, "build")

    # Ensure directories exist
    os.makedirs(assets_dir, exist_ok=True)
    os.makedirs(build_dir, exist_ok=True)

    # All icon sizes needed:
    # - Windows: 16, 20, 24, 32, 40, 48, 64, 128, 256
    # - macOS: 16, 32, 64, 128, 256, 512, 1024
    # - Linux: 16, 22, 24, 32, 36, 48, 64, 72, 96, 128, 192, 256, 512
    # Combined unique sizes for all platforms
    all_sizes = sorted(set([
        16, 20, 22, 24, 32, 36, 40, 48, 64, 72, 96, 128, 192, 256, 512, 1024
    ]))

    icons = {}
    print("Generating MyReviser icons for all platforms...")
    print(f"Sizes: {all_sizes}")
    print()

    # Generate all sizes
    for size in all_sizes:
        icon = create_image(
            size,
            size,
            primary_color="#4A90E2",  # Bright blue
            accent_color="#2C3E50",  # Dark blue
        )
        icons[size] = icon

        # Save individual PNG files (useful for Linux)
        output_file = os.path.join(assets_dir, f"icon_{size}.png")
        icon.save(output_file, format="PNG")
        print(f"  ✓ Created icon_{size}.png")

    # Save main icon.png (256x256) - standard size
    main_icon = os.path.join(assets_dir, "icon.png")
    icons[256].save(main_icon, format="PNG")
    print(f"\n  ✓ Created main icon.png (256x256)")

    # Save the build/appicon.png for Wails (highest quality)
    build_icon = os.path.join(build_dir, "appicon.png")
    icons[1024].save(build_icon, format="PNG")
    print(f"  ✓ Created build/appicon.png (1024x1024 for Wails)")

    # Save ICO file for Windows with multiple sizes
    ico_file = os.path.join(assets_dir, "icon.ico")
    # Windows ICO recommended sizes (ICO format max is 256x256)
    win_sizes = [16, 24, 32, 48, 64, 128, 256]

    # Create ICO with all sizes properly embedded
    # The first image should be the largest for best quality
    icons[256].save(
        ico_file,
        format="ICO",
        sizes=[(size, size) for size in win_sizes]
    )
    print(f"  ✓ Created icon.ico (Windows, multi-resolution: {win_sizes})")

    # Create special Linux sizes that might be needed
    linux_special = {
        'icon_22.png': 22,   # Some older Ubuntu versions
        'icon_36.png': 36,   # Some desktop environments
        'icon_72.png': 72,   # Retina displays
        'icon_96.png': 96,   # Large icons
        'icon_192.png': 192, # Extra large icons
    }

    print("\n  Special Linux sizes:")
    for filename, size in linux_special.items():
        if size in icons:
            print(f"    ✓ {filename} already created")

    print("\n✨ Icon generation complete!")
    print("\nGenerated icons for:")
    print("  • Windows: ICO with sizes 16, 24, 32, 48, 64, 128, 256")
    print("  • macOS: Will use appicon.png to generate ICNS")
    print("  • Linux: PNG files in all standard sizes")

    print("\n📋 Next steps:")
    print("1. Run: cd .. && wails3 generate icons -input build/appicon.png")
    print("2. The ICNS file will be generated at: build/darwin/icons.icns")
    print("3. Copy to assets: cp build/darwin/icons.icns assets/icon.icns")
    print("4. Build the application")