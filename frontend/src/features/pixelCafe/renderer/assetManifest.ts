import lobbySceneUrl from '../assets/scenes/pixel-cafe-lobby-v2.png'
import avatarGoldUrl from '../assets/sprites/avatar-gold.png'
import avatarGoldWalk0Url from '../assets/sprites/avatar-gold-walk-0.png'
import avatarGoldWalk1Url from '../assets/sprites/avatar-gold-walk-1.png'
import avatarGoldWalk2Url from '../assets/sprites/avatar-gold-walk-2.png'
import avatarGoldWalk3Url from '../assets/sprites/avatar-gold-walk-3.png'
import avatarTealUrl from '../assets/sprites/avatar-teal.png'
import avatarTealWalk0Url from '../assets/sprites/avatar-teal-walk-0.png'
import avatarTealWalk1Url from '../assets/sprites/avatar-teal-walk-1.png'
import avatarTealWalk2Url from '../assets/sprites/avatar-teal-walk-2.png'
import avatarTealWalk3Url from '../assets/sprites/avatar-teal-walk-3.png'
import avatarWineUrl from '../assets/sprites/avatar-wine.png'
import avatarWineWalk0Url from '../assets/sprites/avatar-wine-walk-0.png'
import avatarWineWalk1Url from '../assets/sprites/avatar-wine-walk-1.png'
import avatarWineWalk2Url from '../assets/sprites/avatar-wine-walk-2.png'
import avatarWineWalk3Url from '../assets/sprites/avatar-wine-walk-3.png'
import workstationBlueUrl from '../assets/sprites/workstation-blue.png'
import workstationTealUrl from '../assets/sprites/workstation-teal.png'

const WALK_FRAME_ASPECT_RATIO = 96 / 128

export const cafeSceneAssets = {
  lobbyBackground: lobbySceneUrl,
  avatars: [
    {
      url: avatarTealUrl,
      aspectRatio: 53 / 128,
      walkFrameAspectRatio: WALK_FRAME_ASPECT_RATIO,
      walkFrames: [avatarTealWalk0Url, avatarTealWalk1Url, avatarTealWalk2Url, avatarTealWalk3Url],
    },
    {
      url: avatarGoldUrl,
      aspectRatio: 62 / 128,
      walkFrameAspectRatio: WALK_FRAME_ASPECT_RATIO,
      walkFrames: [avatarGoldWalk0Url, avatarGoldWalk1Url, avatarGoldWalk2Url, avatarGoldWalk3Url],
    },
    {
      url: avatarWineUrl,
      aspectRatio: 63 / 128,
      walkFrameAspectRatio: WALK_FRAME_ASPECT_RATIO,
      walkFrames: [avatarWineWalk0Url, avatarWineWalk1Url, avatarWineWalk2Url, avatarWineWalk3Url],
    },
  ],
  workstations: [
    { url: workstationTealUrl, aspectRatio: 192 / 182 },
    { url: workstationBlueUrl, aspectRatio: 192 / 188 },
  ],
} as const
