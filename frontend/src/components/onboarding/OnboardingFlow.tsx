import { useState, KeyboardEvent } from "react";
import { useNavigate } from "react-router-dom";
import { Tag, Globe, Rss, MessageCircle, Check, ArrowRight, Rocket, Shield } from "lucide-react";
import { X } from "lucide-react";
import { cn } from "@/lib/utils";
import { useCreateBrand } from "@/hooks/use-brands";
import { useCreateSource, useStartCollectorJob } from "@/hooks/use-sources";
import { DEFAULT_PROJECT_ID } from "@/lib/constants";
import { toast } from "sonner";

const steps = [
  { num: 1, label: "Бренд", icon: Tag },
  { num: 2, label: "Источники", icon: Globe },
  { num: 3, label: "Сбор", icon: Rocket },
];

function TagInput({
  label,
  hint,
  tags,
  onChange,
  variant = "default",
}: {
  label: string;
  hint?: string;
  tags: string[];
  onChange: (t: string[]) => void;
  variant?: "default" | "risk" | "muted";
}) {
  const [input, setInput] = useState("");

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && input.trim()) {
      e.preventDefault();
      if (!tags.includes(input.trim())) onChange([...tags, input.trim()]);
      setInput("");
    }
  };

  const chipClass =
    variant === "risk"
      ? "bg-destructive/15 text-destructive"
      : variant === "muted"
        ? "bg-secondary text-muted-foreground"
        : "bg-violet-500/15 text-violet-400";

  return (
    <div className="space-y-1.5">
      <label className="text-sm font-medium text-foreground">{label}</label>
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
      {tags.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {tags.map((tag) => (
            <span key={tag} className={cn("text-xs rounded-full px-2.5 py-1 flex items-center gap-1", chipClass)}>
              {tag}
              <button onClick={() => onChange(tags.filter((t) => t !== tag))} className="hover:opacity-70 cursor-pointer"><X className="h-3 w-3" /></button>
            </span>
          ))}
        </div>
      )}
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Введите и нажмите Enter"
        className="w-full bg-secondary border border-border rounded-lg px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-primary"
      />
    </div>
  );
}

const sourceTypes = [
  { type: "Web" as const, icon: Globe, title: "Web", desc: "Парсинг веб-страниц и форумов" },
  { type: "RSS" as const, icon: Rss, title: "RSS", desc: "Мониторинг RSS-лент" },
  { type: "Telegram" as const, icon: MessageCircle, title: "Telegram", desc: "Отслеживание каналов и чатов" },
];

export function OnboardingFlow({ onComplete }: { onComplete: () => void }) {
  const navigate = useNavigate();
  const [step, setStep] = useState(0); // 0=welcome, 1=brand, 2=source, 3=done

  // Brand form
  const [brandName, setBrandName] = useState("");
  const [keywords, setKeywords] = useState<string[]>([]);
  const [exclusions, setExclusions] = useState<string[]>([]);
  const [riskWords, setRiskWords] = useState<string[]>([]);

  // Source form
  const [sourceType, setSourceType] = useState<"Web" | "RSS" | "Telegram">("Web");
  const [sourceName, setSourceName] = useState("");
  const [sourceUrl, setSourceUrl] = useState("");

  // Created IDs
  const [createdSourceId, setCreatedSourceId] = useState<string | null>(null);

  const createBrand = useCreateBrand();
  const createSource = useCreateSource();
  const startJob = useStartCollectorJob();

  const [brandErrors, setBrandErrors] = useState<{ name?: string; keywords?: string }>({});
  const [sourceErrors, setSourceErrors] = useState<{ name?: string }>({});

  const handleBrandNext = () => {
    const e: typeof brandErrors = {};
    if (!brandName.trim()) e.name = "Обязательное поле";
    if (keywords.length === 0) e.keywords = "Добавьте хотя бы одно ключевое слово";
    setBrandErrors(e);
    if (Object.keys(e).length > 0) return;

    createBrand.mutate(
      { data: { name: brandName.trim(), keywords, exclusions, risk_words: riskWords } },
      {
        onSuccess: () => { toast.success("Бренд создан"); setStep(2); },
        onError: () => { toast.error("Не удалось создать бренд"); setStep(2); },
      }
    );
  };

  const handleSourceNext = () => {
    const e: typeof sourceErrors = {};
    if (!sourceName.trim()) e.name = "Обязательное поле";
    setSourceErrors(e);
    if (Object.keys(e).length > 0) return;
    createSource.mutate(
      { data: { type: sourceType, name: sourceName.trim(), config: sourceUrl } },
      {
        onSuccess: (data: { id: string }) => { toast.success("Источник добавлен"); setCreatedSourceId(data?.id ?? null); setStep(3); },
        onError: () => { toast.error("Не удалось добавить источник"); setStep(3); },
      }
    );
  };

  const handleLaunch = () => {
    if (createdSourceId) {
      startJob.mutate(createdSourceId, {
        onSuccess: () => toast.success("Сбор запущен"),
        onError: () => toast.error("Не удалось запустить сбор"),
      });
    }
    finish();
  };

  const finish = () => {
    localStorage.setItem("onboarding_completed", "true");
    onComplete();
    navigate("/");
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-[70vh] animate-page-enter">
      {/* Stepper */}
      {step > 0 && (
        <div className="flex items-center gap-2 mb-10">
          {steps.map((s, i) => {
            const done = step > s.num;
            const active = step === s.num;
            return (
              <div key={s.num} className="flex items-center gap-2">
                <div className={cn(
                  "w-9 h-9 rounded-full flex items-center justify-center text-sm font-bold transition-all duration-300",
                  done ? "bg-primary text-primary-foreground" : active ? "bg-primary/20 text-primary ring-2 ring-primary" : "bg-secondary text-muted-foreground"
                )}>
                  {done ? <Check className="h-4 w-4" /> : s.num}
                </div>
                <span className={cn("text-sm hidden sm:inline", active ? "text-foreground font-medium" : "text-muted-foreground")}>{s.label}</span>
                {i < steps.length - 1 && <div className={cn("w-8 sm:w-12 h-0.5 rounded-full transition-colors", done ? "bg-primary" : "bg-border")} />}
              </div>
            );
          })}
        </div>
      )}

      {/* Step 0 — Welcome */}
      {step === 0 && (
        <div className="text-center max-w-lg space-y-6">
          <div className="mx-auto w-20 h-20 rounded-2xl bg-primary/15 flex items-center justify-center mb-2">
            <Shield className="h-10 w-10 text-primary" />
          </div>
          <h1 className="text-3xl sm:text-4xl font-bold text-foreground">Добро пожаловать в BrandRadar</h1>
          <p className="text-muted-foreground text-lg">Настройте мониторинг репутации за 3 шага</p>
          <div className="flex justify-center gap-6 pt-2">
            {steps.map((s) => (
              <div key={s.num} className="flex flex-col items-center gap-2">
                <div className="w-12 h-12 rounded-xl bg-secondary flex items-center justify-center">
                  <s.icon className="h-5 w-5 text-muted-foreground" />
                </div>
                <span className="text-xs text-muted-foreground">{s.num}. {s.label}</span>
              </div>
            ))}
          </div>
          <button
            onClick={() => setStep(1)}
            className="mt-4 px-8 py-3 bg-primary text-primary-foreground text-base font-medium rounded-xl hover:bg-primary/90 transition-colors cursor-pointer"
          >
            Начать
          </button>
        </div>
      )}

      {/* Step 1 — Brand */}
      {step === 1 && (
        <div className="w-full max-w-lg space-y-6">
          <div className="text-center space-y-1">
            <h2 className="text-2xl font-bold text-foreground">Создайте бренд</h2>
            <p className="text-sm text-muted-foreground">Укажите название и ключевые слова для мониторинга</p>
          </div>

          <div className="bg-card border border-border rounded-xl p-6 space-y-5">
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">Название бренда *</label>
              <input
                value={brandName}
                onChange={(e) => { setBrandName(e.target.value); if (brandErrors.name) setBrandErrors((p) => ({ ...p, name: undefined })); }}
                placeholder="Например: BrandRadar"
                maxLength={50}
                className={cn("w-full bg-secondary border rounded-lg px-4 py-3 text-base text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary", brandErrors.name ? "border-destructive" : "border-border")}
                autoFocus
              />
              {brandErrors.name && <p className="text-xs text-destructive">{brandErrors.name}</p>}
            </div>
            <TagInput label="Ключевые слова *" hint="Введите слово и нажмите Enter" tags={keywords} onChange={(t) => { setKeywords(t); if (brandErrors.keywords) setBrandErrors((p) => ({ ...p, keywords: undefined })); }} />
            {brandErrors.keywords && <p className="text-xs text-destructive -mt-3">{brandErrors.keywords}</p>}
            <TagInput label="Слова-исключения" hint="Опционально" tags={exclusions} onChange={setExclusions} variant="muted" />
            <TagInput label="Risk-слова" hint="Опционально" tags={riskWords} onChange={setRiskWords} variant="risk" />
          </div>

          <div className="flex justify-end">
            <button
              onClick={handleBrandNext}
              disabled={!brandName.trim() || createBrand.isPending}
              className="px-6 py-2.5 bg-primary text-primary-foreground text-sm font-medium rounded-xl hover:bg-primary/90 transition-colors disabled:opacity-50 flex items-center gap-2 cursor-pointer"
            >
              {createBrand.isPending ? "Создание..." : <>Далее <ArrowRight className="h-4 w-4" /></>}
            </button>
          </div>
        </div>
      )}

      {/* Step 2 — Source */}
      {step === 2 && (
        <div className="w-full max-w-lg space-y-6">
          <div className="text-center space-y-1">
            <h2 className="text-2xl font-bold text-foreground">Добавьте источник</h2>
            <p className="text-sm text-muted-foreground">Откуда собирать упоминания?</p>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {sourceTypes.map((st) => (
              <button
                key={st.type}
                onClick={() => setSourceType(st.type)}
                className={cn(
                  "bg-card border rounded-xl p-4 flex flex-col items-center gap-2 transition-all duration-200 cursor-pointer text-center",
                  sourceType === st.type ? "border-primary ring-2 ring-primary/30" : "border-border hover:border-muted-foreground/40"
                )}
              >
                <st.icon className={cn("h-8 w-8", sourceType === st.type ? "text-primary" : "text-muted-foreground")} />
                <span className={cn("text-sm font-medium", sourceType === st.type ? "text-foreground" : "text-muted-foreground")}>{st.title}</span>
                <span className="text-[11px] text-muted-foreground/60">{st.desc}</span>
              </button>
            ))}
          </div>

          <div className="bg-card border border-border rounded-xl p-6 space-y-4">
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">Название источника *</label>
              <input
                value={sourceName}
                onChange={(e) => { setSourceName(e.target.value); if (sourceErrors.name) setSourceErrors({}); }}
                placeholder="Например: Tech News RSS"
                className={cn("w-full bg-secondary border rounded-lg px-4 py-3 text-base text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary", sourceErrors.name ? "border-destructive" : "border-border")}
              />
              {sourceErrors.name && <p className="text-xs text-destructive">{sourceErrors.name}</p>}
            </div>
            <div className="space-y-1.5">
              <label className="text-sm font-medium text-foreground">URL / Конфиг</label>
              <input
                value={sourceUrl}
                onChange={(e) => setSourceUrl(e.target.value)}
                placeholder="https://..."
                className="w-full bg-secondary border border-border rounded-lg px-4 py-3 text-base text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
              />
            </div>
          </div>

          <div className="flex justify-between">
            <button onClick={() => setStep(1)} className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground transition-colors cursor-pointer">← Назад</button>
            <button
              onClick={handleSourceNext}
              disabled={!sourceName.trim() || createSource.isPending}
              className="px-6 py-2.5 bg-primary text-primary-foreground text-sm font-medium rounded-xl hover:bg-primary/90 transition-colors disabled:opacity-50 flex items-center gap-2 cursor-pointer"
            >
              {createSource.isPending ? "Добавление..." : <>Далее <ArrowRight className="h-4 w-4" /></>}
            </button>
          </div>
        </div>
      )}

      {/* Step 3 — Done */}
      {step === 3 && (
        <div className="text-center max-w-md space-y-6">
          <div className="mx-auto w-20 h-20 rounded-full bg-success/15 flex items-center justify-center">
            <Check className="h-10 w-10 text-success" />
          </div>
          <h2 className="text-2xl font-bold text-foreground">Всё настроено!</h2>
          <p className="text-muted-foreground">Запускаем первый сбор данных?</p>
          <div className="flex flex-col sm:flex-row gap-3 justify-center pt-2">
            <button
              onClick={handleLaunch}
              disabled={startJob.isPending}
              className="px-6 py-3 bg-primary text-primary-foreground text-sm font-medium rounded-xl hover:bg-primary/90 transition-colors cursor-pointer disabled:opacity-50"
            >
              {startJob.isPending ? "Запуск..." : "🚀 Запустить сбор"}
            </button>
            <button
              onClick={finish}
              className="px-6 py-3 bg-secondary text-foreground text-sm font-medium rounded-xl hover:bg-secondary/80 transition-colors cursor-pointer"
            >
              На дашборд
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
