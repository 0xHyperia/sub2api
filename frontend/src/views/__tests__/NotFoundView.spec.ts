import { mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import NotFoundView from '@/views/NotFoundView.vue'

const { authState, backMock, pushMock } = vi.hoisted(() => ({
  authState: { isAuthenticated: false, isAdmin: false },
  backMock: vi.fn(),
  pushMock: vi.fn(),
}))

vi.mock('vue-router', async () => {
  const actual = await vi.importActual<typeof import('vue-router')>('vue-router')
  return {
    ...actual,
    useRouter: () => ({ back: backMock, push: pushMock }),
  }
})

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => authState,
}))

function mountView() {
  return mount(NotFoundView, {
    global: {
      stubs: {
        RouterLink: {
          props: ['to'],
          template: '<a :href="to"><slot /></a>',
        },
        Icon: true,
      },
    },
  })
}

describe('NotFoundView', () => {
  beforeEach(() => {
    authState.isAuthenticated = false
    authState.isAdmin = false
    backMock.mockReset()
    pushMock.mockReset()
  })

  it('renders an internationalized neutral state with a working home fallback', async () => {
    const wrapper = mountView()

    expect(wrapper.get('h1').text()).toBe('errors.pageNotFound')
    expect(wrapper.text()).toContain('errors.pageNotFoundDescription')
    expect(wrapper.get('a[href="/home"]').text()).toContain('errors.backToHome')
    expect(wrapper.find('a[href="#"]').exists()).toBe(false)
    expect(wrapper.html()).not.toContain('bg-gradient')

    await wrapper.get('button').trigger('click')
    expect(pushMock).toHaveBeenCalledWith('/home')
  })

  it('routes authenticated administrators to the admin dashboard', () => {
    authState.isAuthenticated = true
    authState.isAdmin = true

    const wrapper = mountView()
    expect(wrapper.get('a[href="/admin/dashboard"]').text()).toContain('home.goToDashboard')
  })
})
