/**
 * Gas Town Worker Sprite Animation System
 *
 * Animated sprites for polecat workers (Mad Max style characters)
 * Supports idle, walking, and working animation states
 * Includes pathfinding and movement between buildings
 */

/**
 * AnimatedSprite extends Sprite with frame-based animation support
 */
class AnimatedSprite extends Sprite {
  /**
   * @param {Object} options - Sprite configuration
   * @param {Array<Object>} options.frames - Animation frames [{region, duration}, ...]
   * @param {boolean} [options.loop=true] - Whether animation loops
   * @param {boolean} [options.autoPlay=true] - Start playing automatically
   */
  constructor(options) {
    super(options);

    this.frames = options.frames || [];
    this.currentFrame = 0;
    this.frameTime = 0;
    this.loop = options.loop !== undefined ? options.loop : true;
    this.playing = options.autoPlay !== undefined ? options.autoPlay : true;
    this.onComplete = options.onComplete || null;

    // Set initial region from first frame
    if (this.frames.length > 0) {
      this.region = this.frames[0].region;
    }
  }

  /**
   * Update animation
   * @param {number} deltaTime - Time since last update (ms)
   */
  update(deltaTime) {
    if (!this.playing || this.frames.length === 0) return;

    this.frameTime += deltaTime;
    const currentFrameData = this.frames[this.currentFrame];

    if (this.frameTime >= currentFrameData.duration) {
      this.frameTime = 0;
      this.currentFrame++;

      if (this.currentFrame >= this.frames.length) {
        if (this.loop) {
          this.currentFrame = 0;
        } else {
          this.currentFrame = this.frames.length - 1;
          this.playing = false;
          if (this.onComplete) {
            this.onComplete();
          }
        }
      }

      // Update sprite region to current frame
      this.region = this.frames[this.currentFrame].region;
    }
  }

  /**
   * Play the animation
   */
  play() {
    this.playing = true;
  }

  /**
   * Pause the animation
   */
  pause() {
    this.playing = false;
  }

  /**
   * Reset animation to first frame
   */
  reset() {
    this.currentFrame = 0;
    this.frameTime = 0;
    if (this.frames.length > 0) {
      this.region = this.frames[0].region;
    }
  }

  /**
   * Set animation frames
   * @param {Array<Object>} frames - New animation frames
   * @param {boolean} [resetToStart=true] - Reset to first frame
   */
  setFrames(frames, resetToStart = true) {
    this.frames = frames;
    if (resetToStart) {
      this.reset();
    }
  }
}

/**
 * WorkerState enum
 */
const WorkerState = {
  IDLE: 'idle',
  WALKING: 'walking',
  WORKING: 'working'
};

/**
 * WorkerSprite manages a polecat worker with animations and movement
 */
class WorkerSprite extends AnimatedSprite {
  /**
   * @param {Object} options - Worker configuration
   * @param {TextureAtlas} options.atlas - Texture atlas with worker sprites
   * @param {Object} options.animations - Map of state to frame arrays
   * @param {string} [options.name='Worker'] - Worker name
   * @param {number} [options.speed=50] - Movement speed (pixels/second)
   */
  constructor(options) {
    // Start with idle animation
    const idleFrames = options.animations[WorkerState.IDLE] || [];

    super({
      ...options,
      texture: options.atlas,
      frames: idleFrames,
      loop: true,
      autoPlay: true
    });

    this.name = options.name || 'Worker';
    this.animations = options.animations;
    this.state = WorkerState.IDLE;
    this.speed = options.speed || 50; // pixels per second

    // Movement
    this.targetX = null;
    this.targetY = null;
    this.path = [];
    this.currentPathIndex = 0;

    // Work tracking
    this.currentBuilding = null;
    this.targetBuilding = null;
  }

  /**
   * Set animation state
   * @param {string} newState - WorkerState value
   */
  setState(newState) {
    if (this.state === newState) return;

    this.state = newState;
    const frames = this.animations[newState];

    if (frames) {
      this.setFrames(frames, true);
      this.play();
    }
  }

  /**
   * Move to a target position
   * @param {number} x - Target X coordinate
   * @param {number} y - Target Y coordinate
   */
  moveTo(x, y) {
    this.targetX = x;
    this.targetY = y;
    this.setState(WorkerState.WALKING);
  }

  /**
   * Move to a building
   * @param {Object} building - Building object with x, y coordinates
   */
  moveToBuilding(building) {
    this.targetBuilding = building;
    this.moveTo(building.x, building.y);
  }

  /**
   * Update worker position and animation
   * @param {number} deltaTime - Time since last update (ms)
   */
  update(deltaTime) {
    // Update animation
    super.update(deltaTime);

    // Update movement
    if (this.targetX !== null && this.targetY !== null) {
      const dx = this.targetX - this.x;
      const dy = this.targetY - this.y;
      const distance = Math.sqrt(dx * dx + dy * dy);

      // Arrived at target
      if (distance < 5) {
        this.x = this.targetX;
        this.y = this.targetY;
        this.targetX = null;
        this.targetY = null;

        // If we reached a building, start working
        if (this.targetBuilding) {
          this.currentBuilding = this.targetBuilding;
          this.targetBuilding = null;
          this.setState(WorkerState.WORKING);

          // After working for a while, go back to idle
          setTimeout(() => {
            this.setState(WorkerState.IDLE);
          }, 2000 + Math.random() * 3000); // 2-5 seconds
        } else {
          this.setState(WorkerState.IDLE);
        }
      } else {
        // Move toward target
        const moveDistance = (this.speed * deltaTime) / 1000;
        const ratio = moveDistance / distance;

        this.x += dx * ratio;
        this.y += dy * ratio;

        // Update rotation to face movement direction
        this.rotation = Math.atan2(dy, dx);
      }
    }
  }

  /**
   * Scurry to a random nearby position (for idle wandering)
   */
  scurry() {
    const angle = Math.random() * Math.PI * 2;
    const distance = 20 + Math.random() * 80; // 20-100 pixels
    const x = this.x + Math.cos(angle) * distance;
    const y = this.y + Math.sin(angle) * distance;
    this.moveTo(x, y);
  }
}

/**
 * WorkerManager manages all worker sprites and syncs with polecat data
 */
class WorkerManager {
  /**
   * @param {TextureAtlas} atlas - Texture atlas for worker sprites
   * @param {Object} animations - Animation definitions
   * @param {Array<Object>} buildings - Building positions
   */
  constructor(atlas, animations, buildings = []) {
    this.atlas = atlas;
    this.animations = animations;
    this.buildings = buildings;
    this.workers = new Map(); // Map of polecat name -> WorkerSprite
    this.spriteRenderer = new SpriteRenderer();
  }

  /**
   * Create or update a worker sprite from polecat data
   * @param {Object} polecatData - Polecat info {name, rig, status, etc.}
   */
  updateWorker(polecatData) {
    const workerId = `${polecatData.rig}/${polecatData.name}`;

    let worker = this.workers.get(workerId);

    if (!worker) {
      // Create new worker
      // Start at a random position or assigned building
      const startX = (Math.random() - 0.5) * 400;
      const startY = (Math.random() - 0.5) * 400;

      worker = new WorkerSprite({
        atlas: this.atlas,
        animations: this.animations,
        name: polecatData.name,
        x: startX,
        y: startY,
        zIndex: 10
      });

      this.workers.set(workerId, worker);
      this.spriteRenderer.add(worker);

      // Randomly send worker to a building
      if (this.buildings.length > 0) {
        const randomBuilding = this.buildings[Math.floor(Math.random() * this.buildings.length)];
        worker.moveToBuilding(randomBuilding);
      }
    }

    // Update worker based on polecat status
    // You can enhance this to map polecat activity to worker states

    return worker;
  }

  /**
   * Remove a worker sprite
   * @param {string} polecatName - Polecat identifier
   */
  removeWorker(polecatName) {
    const worker = this.workers.get(polecatName);
    if (worker) {
      this.spriteRenderer.remove(worker);
      this.workers.delete(polecatName);
    }
  }

  /**
   * Update all workers
   * @param {number} deltaTime - Time since last update (ms)
   */
  update(deltaTime) {
    for (const worker of this.workers.values()) {
      worker.update(deltaTime);

      // Randomly make idle workers scurry around
      if (worker.state === WorkerState.IDLE && Math.random() < 0.002) {
        worker.scurry();
      }
    }
  }

  /**
   * Render all workers
   * @param {CanvasRenderingContext2D} ctx
   */
  render(ctx) {
    this.spriteRenderer.render(ctx);
  }

  /**
   * Sync workers with polecat data from API
   * @param {Array<Object>} polecatData - Array of polecat info
   */
  syncWithPolecats(polecatData) {
    const activeWorkers = new Set();

    for (const polecat of polecatData) {
      const workerId = `${polecat.rig}/${polecat.name}`;
      this.updateWorker(polecat);
      activeWorkers.add(workerId);
    }

    // Remove workers that no longer exist
    for (const workerId of this.workers.keys()) {
      if (!activeWorkers.has(workerId)) {
        this.removeWorker(workerId);
      }
    }
  }
}

/**
 * Create default Mad Max worker animations
 * This creates placeholder animations - replace with actual sprite regions
 * @param {TextureAtlas} atlas - Texture atlas
 * @returns {Object} Animation definitions by state
 */
function createDefaultWorkerAnimations(atlas) {
  // Placeholder animations - will need actual sprite regions
  // For now, using a single frame for each state
  return {
    [WorkerState.IDLE]: [
      { region: 'worker_idle_0', duration: 200 },
      { region: 'worker_idle_1', duration: 200 },
      { region: 'worker_idle_0', duration: 200 },
    ],
    [WorkerState.WALKING]: [
      { region: 'worker_walk_0', duration: 100 },
      { region: 'worker_walk_1', duration: 100 },
      { region: 'worker_walk_2', duration: 100 },
      { region: 'worker_walk_3', duration: 100 },
    ],
    [WorkerState.WORKING]: [
      { region: 'worker_work_0', duration: 150 },
      { region: 'worker_work_1', duration: 150 },
      { region: 'worker_work_2', duration: 150 },
    ]
  };
}

/**
 * Create a simple worker texture atlas from a single color
 * This is a placeholder - replace with actual Mad Max sprite sheet
 * @param {string} color - Worker color
 * @returns {Promise<TextureAtlas>}
 */
async function createPlaceholderWorkerAtlas(color = '#f87171') {
  // Create a simple canvas-based sprite sheet
  const canvas = document.createElement('canvas');
  canvas.width = 128;
  canvas.height = 128;
  const ctx = canvas.getContext('2d');

  // Helper to draw simple character shape
  function drawWorker(x, y, pose = 0) {
    ctx.fillStyle = color;

    // Body (rectangle)
    ctx.fillRect(x + 6, y + 12, 12, 16);

    // Head (circle)
    ctx.beginPath();
    ctx.arc(x + 12, y + 8, 6, 0, Math.PI * 2);
    ctx.fill();

    // Legs (based on pose)
    if (pose === 0) {
      ctx.fillRect(x + 7, y + 28, 4, 8);
      ctx.fillRect(x + 13, y + 28, 4, 8);
    } else {
      ctx.fillRect(x + 7, y + 28, 4, 6);
      ctx.fillRect(x + 13, y + 30, 4, 6);
    }
  }

  // Draw idle frames (same pose, slight variation)
  drawWorker(0, 0, 0);    // idle_0
  drawWorker(24, 0, 0);   // idle_1

  // Draw walking frames (alternating leg positions)
  drawWorker(48, 0, 0);   // walk_0
  drawWorker(72, 0, 1);   // walk_1
  drawWorker(96, 0, 0);   // walk_2
  drawWorker(0, 40, 1);   // walk_3

  // Draw working frames (different poses)
  drawWorker(24, 40, 1);  // work_0
  drawWorker(48, 40, 0);  // work_1
  drawWorker(72, 40, 1);  // work_2

  // Convert canvas to image
  const dataURL = canvas.toDataURL();
  const image = await loadImage(dataURL);

  // Define regions
  const regions = {
    'worker_idle_0': { x: 0, y: 0, width: 24, height: 36 },
    'worker_idle_1': { x: 24, y: 0, width: 24, height: 36 },

    'worker_walk_0': { x: 48, y: 0, width: 24, height: 36 },
    'worker_walk_1': { x: 72, y: 0, width: 24, height: 36 },
    'worker_walk_2': { x: 96, y: 0, width: 24, height: 36 },
    'worker_walk_3': { x: 0, y: 40, width: 24, height: 36 },

    'worker_work_0': { x: 24, y: 40, width: 24, height: 36 },
    'worker_work_1': { x: 48, y: 40, width: 24, height: 36 },
    'worker_work_2': { x: 72, y: 40, width: 24, height: 36 },
  };

  return new TextureAtlas(image, regions);
}

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    AnimatedSprite,
    WorkerState,
    WorkerSprite,
    WorkerManager,
    createDefaultWorkerAnimations,
    createPlaceholderWorkerAtlas
  };
}
