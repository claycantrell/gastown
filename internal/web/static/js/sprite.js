/**
 * Gas Town 2D Sprite Rendering System
 *
 * Coordinate system: Origin at center, +x right, +y down
 * Provides sprite rendering with texture atlas support and z-index layering
 */

/**
 * TextureAtlas manages sprite sheets and texture regions
 */
class TextureAtlas {
  /**
   * @param {HTMLImageElement} image - The sprite sheet image
   * @param {Object} regions - Map of region names to {x, y, width, height}
   */
  constructor(image, regions = {}) {
    this.image = image;
    this.regions = regions;
  }

  /**
   * Add a texture region to the atlas
   * @param {string} name - Region identifier
   * @param {number} x - X coordinate in sprite sheet
   * @param {number} y - Y coordinate in sprite sheet
   * @param {number} width - Region width
   * @param {number} height - Region height
   */
  addRegion(name, x, y, width, height) {
    this.regions[name] = { x, y, width, height };
  }

  /**
   * Get a texture region by name
   * @param {string} name - Region identifier
   * @returns {Object|null} Region data or null if not found
   */
  getRegion(name) {
    return this.regions[name] || null;
  }

  /**
   * Create atlas from a grid layout
   * @param {HTMLImageElement} image - The sprite sheet image
   * @param {number} spriteWidth - Width of each sprite
   * @param {number} spriteHeight - Height of each sprite
   * @param {number} columns - Number of columns in grid
   * @param {number} rows - Number of rows in grid
   * @param {Array<string>} names - Names for each sprite (row-major order)
   * @returns {TextureAtlas}
   */
  static fromGrid(image, spriteWidth, spriteHeight, columns, rows, names = []) {
    const regions = {};
    let nameIndex = 0;

    for (let row = 0; row < rows; row++) {
      for (let col = 0; col < columns; col++) {
        const name = names[nameIndex] || `sprite_${nameIndex}`;
        regions[name] = {
          x: col * spriteWidth,
          y: row * spriteHeight,
          width: spriteWidth,
          height: spriteHeight
        };
        nameIndex++;
      }
    }

    return new TextureAtlas(image, regions);
  }
}

/**
 * Sprite represents a drawable 2D sprite
 */
class Sprite {
  /**
   * @param {Object} options - Sprite configuration
   * @param {number} options.x - X position (world coordinates)
   * @param {number} options.y - Y position (world coordinates)
   * @param {HTMLImageElement|TextureAtlas} options.texture - Image or atlas
   * @param {string} [options.region] - Region name if using atlas
   * @param {number} [options.rotation=0] - Rotation in radians
   * @param {number} [options.scale=1] - Uniform scale factor
   * @param {number} [options.scaleX=1] - X-axis scale factor
   * @param {number} [options.scaleY=1] - Y-axis scale factor
   * @param {number} [options.zIndex=0] - Layer depth (higher = front)
   * @param {number} [options.alpha=1] - Opacity (0-1)
   * @param {number} [options.anchorX=0.5] - X anchor point (0-1)
   * @param {number} [options.anchorY=0.5] - Y anchor point (0-1)
   */
  constructor(options) {
    this.x = options.x || 0;
    this.y = options.y || 0;
    this.texture = options.texture;
    this.region = options.region || null;
    this.rotation = options.rotation || 0;
    this.scale = options.scale || 1;
    this.scaleX = options.scaleX !== undefined ? options.scaleX : this.scale;
    this.scaleY = options.scaleY !== undefined ? options.scaleY : this.scale;
    this.zIndex = options.zIndex || 0;
    this.alpha = options.alpha !== undefined ? options.alpha : 1;
    this.anchorX = options.anchorX !== undefined ? options.anchorX : 0.5;
    this.anchorY = options.anchorY !== undefined ? options.anchorY : 0.5;
    this.visible = true;
  }

  /**
   * Get texture dimensions
   * @returns {Object} {width, height}
   */
  getDimensions() {
    if (this.texture instanceof TextureAtlas && this.region) {
      const region = this.texture.getRegion(this.region);
      return region ? { width: region.width, height: region.height } : { width: 0, height: 0 };
    } else if (this.texture instanceof HTMLImageElement) {
      return { width: this.texture.width, height: this.texture.height };
    }
    return { width: 0, height: 0 };
  }

  /**
   * Draw the sprite
   * @param {CanvasRenderingContext2D} ctx - Canvas context
   */
  draw(ctx) {
    if (!this.visible || this.alpha <= 0) return;

    ctx.save();

    // Apply alpha
    ctx.globalAlpha = this.alpha;

    // Translate to sprite position
    ctx.translate(this.x, this.y);

    // Apply rotation
    if (this.rotation !== 0) {
      ctx.rotate(this.rotation);
    }

    // Apply scale
    ctx.scale(this.scaleX, this.scaleY);

    // Get dimensions
    const dims = this.getDimensions();
    const anchorOffsetX = -dims.width * this.anchorX;
    const anchorOffsetY = -dims.height * this.anchorY;

    // Draw from texture atlas or image
    if (this.texture instanceof TextureAtlas && this.region) {
      const region = this.texture.getRegion(this.region);
      if (region) {
        ctx.drawImage(
          this.texture.image,
          region.x, region.y, region.width, region.height,
          anchorOffsetX, anchorOffsetY, region.width, region.height
        );
      }
    } else if (this.texture instanceof HTMLImageElement) {
      ctx.drawImage(
        this.texture,
        anchorOffsetX, anchorOffsetY,
        dims.width, dims.height
      );
    }

    ctx.restore();
  }
}

/**
 * SpriteRenderer manages sprite rendering with layer support
 */
class SpriteRenderer {
  constructor() {
    this.sprites = [];
  }

  /**
   * Add a sprite to the renderer
   * @param {Sprite} sprite
   */
  add(sprite) {
    this.sprites.push(sprite);
    this.sortByZIndex();
  }

  /**
   * Remove a sprite from the renderer
   * @param {Sprite} sprite
   */
  remove(sprite) {
    const index = this.sprites.indexOf(sprite);
    if (index !== -1) {
      this.sprites.splice(index, 1);
    }
  }

  /**
   * Clear all sprites
   */
  clear() {
    this.sprites = [];
  }

  /**
   * Sort sprites by z-index (back to front)
   */
  sortByZIndex() {
    this.sprites.sort((a, b) => a.zIndex - b.zIndex);
  }

  /**
   * Render all sprites
   * @param {CanvasRenderingContext2D} ctx
   */
  render(ctx) {
    for (const sprite of this.sprites) {
      sprite.draw(ctx);
    }
  }
}

/**
 * Global sprite renderer instance
 */
let globalRenderer = null;

/**
 * Initialize the sprite system with a canvas context
 * @param {CanvasRenderingContext2D} ctx - Canvas context
 * @returns {SpriteRenderer}
 */
function initSpriteSystem(ctx) {
  if (!globalRenderer) {
    globalRenderer = new SpriteRenderer();
  }
  return globalRenderer;
}

/**
 * Convenience function to draw a sprite
 * @param {CanvasRenderingContext2D} ctx - Canvas context
 * @param {number} x - X position
 * @param {number} y - Y position
 * @param {HTMLImageElement|TextureAtlas} texture - Texture
 * @param {number} [rotation=0] - Rotation in radians
 * @param {number} [scale=1] - Scale factor
 * @param {Object} [options={}] - Additional sprite options
 */
function drawSprite(ctx, x, y, texture, rotation = 0, scale = 1, options = {}) {
  const sprite = new Sprite({
    x,
    y,
    texture,
    rotation,
    scale,
    ...options
  });
  sprite.draw(ctx);
}

/**
 * Load an image and return a promise
 * @param {string} src - Image URL
 * @returns {Promise<HTMLImageElement>}
 */
function loadImage(src) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = reject;
    img.src = src;
  });
}

/**
 * Load a texture atlas from an image and region data
 * @param {string} imageSrc - Sprite sheet URL
 * @param {Object} regions - Region definitions
 * @returns {Promise<TextureAtlas>}
 */
async function loadAtlas(imageSrc, regions) {
  const image = await loadImage(imageSrc);
  return new TextureAtlas(image, regions);
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    Sprite,
    TextureAtlas,
    SpriteRenderer,
    initSpriteSystem,
    drawSprite,
    loadImage,
    loadAtlas
  };
}
