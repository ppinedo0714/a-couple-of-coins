import { useSearchParams } from 'react-router-dom'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { PageWrapper } from '@/components/layout/PageWrapper'
import { AccountsTab } from './AccountsTab'
import { CategoriesTab } from './CategoriesTab'
import { ProfileTab } from './ProfileTab'

const TABS = ['accounts', 'categories', 'profile'] as const
type TabKey = (typeof TABS)[number]

export default function SettingsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const raw = searchParams.get('tab')
  const tab: TabKey = (TABS as readonly string[]).includes(raw ?? '') ? (raw as TabKey) : 'accounts'

  const setTab = (next: string) => {
    const params = new URLSearchParams(searchParams)
    if (next === 'accounts') params.delete('tab')
    else params.set('tab', next)
    setSearchParams(params, { replace: true })
  }

  return (
    <PageWrapper className="space-y-6">
      <div>
        <h1 className="font-serif text-3xl">Settings</h1>
        <p className="text-sm text-muted-foreground">Manage accounts, categories, and your profile.</p>
      </div>
      <Tabs value={tab} onValueChange={setTab}>
        <TabsList>
          <TabsTrigger value="accounts">Accounts</TabsTrigger>
          <TabsTrigger value="categories">Categories</TabsTrigger>
          <TabsTrigger value="profile">Profile</TabsTrigger>
        </TabsList>
        <TabsContent value="accounts">
          <AccountsTab />
        </TabsContent>
        <TabsContent value="categories">
          <CategoriesTab />
        </TabsContent>
        <TabsContent value="profile">
          <ProfileTab />
        </TabsContent>
      </Tabs>
    </PageWrapper>
  )
}
