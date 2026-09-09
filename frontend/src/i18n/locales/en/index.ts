import landing from './landing'
import common from './common'
import dashboard from './dashboard'
import channelMonitorV2 from './channelMonitorV2'
import batchImage from './batchImage'
import admin from './admin'
import misc from './misc'

export default {
  ...landing,
  ...common,
  ...dashboard,
  ...channelMonitorV2,
  ...batchImage,
  ...misc,
  // `misc` contains a small legacy admin namespace for the onboarding tour
  // and business cards. Merge it instead of letting it replace the complete
  // admin locale imported from ./admin.
  admin: { ...((misc as any).admin ?? {}), ...admin },
}
