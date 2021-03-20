import numpy
import sys

import random
import os 


ENV = numpy.load("./vol6_environment.npy")

SEED=7
xseed = SEED

random.seed(SEED)


def randi(upto):
  return random.randint (0, upto - 1)



TURNS= 2000
AGENTSPERTURN = 100
RANDOMGEN = 3 # Out of 100
RANDOMAGENTMOVE = 10 # Out of 100
XSIZE, YSIZE, ZSIZE = 300, 300, 1250
shapes = ENV.shape
if XSIZE > shapes[0] or YSIZE > shapes[1] or ZSIZE > shapes[2]:
    raise Exception("Environment file not big enough for specified size")
AGENTS_TRACKED = 10 # Number of agents born every turn, up to all agents born each turn, that are tracked.
if AGENTS_TRACKED > AGENTSPERTURN:
    raise Exception("Agents tracked should be at most agents generated per turn")

agents = []
tracking = numpy.full((AGENTSPERTURN * TURNS, TURNS, 3), -1, dtype=numpy.int16)
track2 = numpy.full(AGENTSPERTURN * TURNS * TURNS * 3, -1, dtype=numpy.int16)

# Probability array. Dimensions (-dz, dy, dx) from agent. Zero means no likelihood. Numbers higher represent the number of tickets the possibility has. One ticket wins. If the winning move is blocked, the process is restarted.
PROBS = [

    [
        [1, 1, 1],
        [1, 0, 1],
        [1, 1, 1],
    ],
    [
        [2, 2, 2],
        [2, 3, 2],
        [2, 2, 2],
    ]
]

# Initialize the cumulative probability array.
CPROBS = []
def initCprobs():
  cur = 0
  for a, x in enumerate(PROBS):
      for b, y in enumerate(x):
          for c, z in enumerate(y):
              if z == 0:
                continue
              cur += z
              CPROBS.append((cur, c - 1, b - 1, -a))






# Move preference. Dimensions (-dz, dy, dx) from agent. Since the z direction is reversed, so is dz. The point at dz = 0, x and y = 0, is ignored. If there is a tie in the calculated preference, one is chosen randomly.
PREFS = [

    [
        [1, 1, 1],
        [1, 0, 1],
        [1, 1, 1],
    ],
    [
        [2, 2, 2],
        [2, 3, 2],
        [2, 2, 2],
    ]
]


# cell
initCprobs()


class Volume:
  def startUp(self):
    self.count = 0
    self.ar = [[[False for a in range(ZSIZE)] for b in range(YSIZE)]for c in range(XSIZE)]
  def pres(self, x, y, z):
    return self.ar[x][y][z]
  def put(self, x,y,z):
    self.ar[x][y][z] = True
    self.count += 1
  def remove(self, x,y,z):
    self.ar[x][y][z] = False
    self.count -= 1
  def full(self):
    return self.count == XSIZE * YSIZE * ZSIZE

volume = Volume()
volume.startUp()

def legal(x, y, z):
    if x in range(XSIZE) and y in range(YSIZE) and z in range(ZSIZE):
        if not volume.pres(x,y,z):
            return True
    return False


class Agent:
    def startUp(self):
      self.x = -1
      self.y = -1
      self.z = -1
      self.dead = False
    def hasRoom(self):
      for dz in range(-1, 1):
        for dy in range(-1, 2):
          for dx in range(-1, 2):
            if dz == 0 and dy == 0 and dx == 0:
              continue
            if legal(self.x + dx, self.y + dy, self.z + dz):
              return True
      return False

    def executeMove(self, dx, dy, dz):
        volume.remove(self.x, self.y, self.z)
        self.x += dx
        self.y += dy
        self.z += dz

        if not legal(self.x, self.y, self.z):
          raise Exception
        volume.put(self.x, self.y, self.z)
        if self.z == 0:
          self.dead = True
    def iter(self):
        if self.dead:
            return
        if randi(100) < RANDOMAGENTMOVE:
             if not self.hasRoom():
                self.dead = True
                return

             self.randMove()
             return
        self.prefMove()
        
    def prefMove(self):
        maxn = -1
        maxes = []
        for ai, aa in enumerate(PREFS):
            for bi, bb in enumerate(aa):
                for ci, pre in enumerate(bb):
                    dx = ci - 1
                    dy = bi - 1
                    dz = -ai
                    if dx == 0 and dy == 0 and dz == 0:
                      continue
                    tx = self.x + dx
                    ty = self.y + dy
                    tz = self.z + dz
                    if not legal(tx, ty, tz):
                        continue
                    v = pre
                    v += ENV[tx][ty][tz]
                    if v >= maxn:
                        maxes.append((dx, dy, dz))
                        maxn = v
        if len(maxes) == 0:
          self.dead = True
          return
        cr = randi(len(maxes))
        choose = maxes[cr]
        (dx, dy, dz) = choose
        self.executeMove(dx, dy, dz)

    def randMove(self):

        while True:
            choice = randi(CPROBS[-1][0])
            for cp in CPROBS:
              if choice < cp[0]:
                g, dx, dy, dz = cp
                if legal(self.x + dx, self.y + dy, self.z + dz):
                  self.executeMove(dx, dy, dz)
                  return
                break
        

agents = []

def process():
    for turn in range(TURNS):
        print("turn ", turn)
        for agent in range(AGENTSPERTURN):
            agent = Agent()
            agent.startUp()
            
            while True:
                if volume.full():
                  raise Exception
                x = randi(XSIZE)
                y = randi(YSIZE)
                if randi(100) < RANDOMGEN:
                  z = randi(ZSIZE)
                else:
                  z = ZSIZE - 1
                if legal(x, y, z):
                   break
            agent.x, agent.y, agent.z = x, y, z
            if agent.z == 0:
              agent.dead = True
            volume.put(agent.x, agent.y, agent.z)
            agents.append(agent)
        count = 0
        for k, agent in enumerate(agents):
           count += 1
           pass
           if k % AGENTSPERTURN < AGENTS_TRACKED:
             tracking[k, turn] = [agent.x, agent.y, agent.z]
           agent.iter()

f = open('sub.csv', 'w')
print("x,y,z", file=f)
process()
for agent in agents:
  print(agent.x, agent.y, agent.z, sep=',', file=f)

f.close()
numpy.save('track.npy', tracking)
