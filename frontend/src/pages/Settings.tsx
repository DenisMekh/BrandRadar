import { usePageTitle } from "@/hooks/usePageTitle";
import { SystemSettings } from "@/components/settings/SystemSettings";

const SettingsPage = () => {
  usePageTitle("Настройки");

  return (
    <div className="space-y-6">
      <h3 className="text-lg font-medium text-foreground">Настройки системы</h3>
      <SystemSettings />
    </div>
  );
};

export default SettingsPage;
