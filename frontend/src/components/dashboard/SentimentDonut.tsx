import { useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { PieChart, Pie, Cell, ResponsiveContainer } from "recharts";
import { fmtNum } from "@/lib/format";
import { useMentions } from "@/hooks/use-mentions";
import { Skeleton } from "@/components/shared/Skeleton";

const sentimentMap: Record<string, string> = {
  "Позитивные": "positive",
  "Негативные": "negative",
  "Нейтральные": "neutral",
};

const COLORS: Record<string, string> = {
  positive: "hsl(142 71% 45.3%)",
  negative: "hsl(0 84.2% 60.2%)",
  neutral: "hsl(240 5% 64.9%)",
};

const LABELS: Record<string, string> = {
  positive: "Позитивные",
  negative: "Негативные",
  neutral: "Нейтральные",
};


export interface SentimentDonutProps {
  total: number;
  positive: number;
  negative: number;
  neutral: number;
  isLoading?: boolean;
  brand_id?: string;
}

export function SentimentDonut({ total, positive, negative, neutral, isLoading, brand_id }: SentimentDonutProps) {
  const navigate = useNavigate();

  const data = useMemo(() => {
    const sum = positive + negative + neutral;
    // Fall back to sampleTotal = 1 to avoid NaN when sum is 0
    const sampleTotal = sum > 0 ? sum : 1;

    return [
      {
        name: LABELS.positive,
        key: "positive",
        value: positive,
        percentage: Math.round((positive / sampleTotal) * 100),
        color: COLORS.positive,
      },
      {
        name: LABELS.negative,
        key: "negative",
        value: negative,
        percentage: Math.round((negative / sampleTotal) * 100),
        color: COLORS.negative,
      },
      {
        name: LABELS.neutral,
        key: "neutral",
        value: neutral,
        percentage: Math.round((neutral / sampleTotal) * 100),
        color: COLORS.neutral,
      }
    ];
  }, [positive, negative, neutral]);

  const handleClick = (_: unknown, index: number) => {
    const segment = data[index];
    if (segment) {
      navigate(`/mentions?sentiment=${segment.key}${brand_id ? `&brand_id=${brand_id}` : ''}`);
    }
  };

  return (
    <>
      <h3 className="text-lg font-medium text-foreground mb-6">Тональность</h3>
      <div className="h-[200px] relative">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={data}
              cx="50%"
              cy="50%"
              innerRadius={60}
              outerRadius={85}
              paddingAngle={3}
              dataKey="value"
              strokeWidth={0}
              isAnimationActive={true}
              animationDuration={800}
              onClick={handleClick}
              style={{ cursor: "pointer" }}
            >
              {data.map((entry, index) => (
                <Cell key={index} fill={entry.color} className="cursor-pointer" />
              ))}
            </Pie>
          </PieChart>
        </ResponsiveContainer>
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <div className="text-center">
            {isLoading ? (
              <Skeleton className="w-12 h-8 mx-auto" />
            ) : (
              <p className="text-2xl font-bold text-foreground">{fmtNum(total)}</p>
            )}
            <p className="text-xs text-muted-foreground">всего</p>
          </div>
        </div>
      </div>
      <div className="flex flex-col gap-2 mt-4">
        {data.map((item) => (
          <button
            key={item.name}
            onClick={() => navigate(`/mentions?sentiment=${item.key}${brand_id ? `&brand_id=${brand_id}` : ''}`)}
            className="flex items-center justify-between text-sm hover:bg-secondary/50 rounded-lg px-2 py-1 -mx-2 transition-colors cursor-pointer"
          >
            <div className="flex items-center gap-2">
              <span className="w-2.5 h-2.5 rounded-full" style={{ background: item.color }} />
              <span className="text-muted-foreground">{item.name}</span>
            </div>
            <span className="text-foreground font-medium">{item.percentage}%</span>
          </button>
        ))}
      </div>
    </>
  );
}
