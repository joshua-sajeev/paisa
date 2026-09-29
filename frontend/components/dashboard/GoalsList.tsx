'use client';

import { DashboardGoal } from '@/lib/dashboard';
import {
  Add,
  LaptopMac,
  FlightTakeoff,
  Shield,
} from '@material-symbols-svg/react/w400';

interface GoalsListProps {
  goals: DashboardGoal[];
  onAddGoal?: () => void;
}

const GOAL_STYLES = [
  {
    background: '#067FF3',
    iconColor: '#FFFFFF',
    progressColor: '#07B682',
    badgeBackground: '#E6F7F2',
    badgeColor: '#07B682',
    icon: LaptopMac,
  },
  {
    background: '#FFB7F1',
    iconColor: '#831843',
    progressColor: '#F7CD1B',
    badgeBackground: '#FFFBEB',
    badgeColor: '#B45309',
    icon: FlightTakeoff,
  },
  {
    background: '#F7CD1B',
    iconColor: '#171717',
    progressColor: '#07B682',
    badgeBackground: '#E6F7F2',
    badgeColor: '#07B682',
    icon: Shield,
  },
];

export default function GoalsList({
  goals,
  onAddGoal,
}: GoalsListProps) {
  return (
    <section className="flex flex-col gap-2.5">
      <div className="flex items-center justify-between">
        <h2 className="text-[15px] font-bold text-slate-900 tracking-tight">
          Financial Goals
        </h2>

        <button
          type="button"
          onClick={onAddGoal}
          className="text-xs font-semibold text-[#067FF3] flex items-center gap-0.5 cursor-pointer"
        >
          <Add className="w-3.5 h-3.5" />
          <span>Add Goal</span>
        </button>
      </div>

      <div className="flex flex-col gap-2">
        {goals.map((goal, index) => {
          const style = GOAL_STYLES[index % GOAL_STYLES.length];
          const Icon = style.icon;

          const progress = Math.min(Math.max(goal.progress, 0), 100);
          const remainingMonths = getRemainingMonths(goal.deadline);

          return (
            <div
              key={goal.id}
              className="bg-white border border-gray-200/80 rounded-xl p-3 flex flex-col gap-2 shadow-sm transition-all hover:shadow-md"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2.5">
                  <div
                    className="w-8 h-8 rounded-lg flex items-center justify-center flex-shrink-0"
                    style={{
                      backgroundColor: style.background,
                      color: style.iconColor,
                    }}
                  >
                    <Icon className="w-4 h-4" />
                  </div>

                  <div className="flex flex-col">
                    <span className="text-xs font-bold text-slate-900 leading-tight">
                      {goal.name}
                    </span>

                    {remainingMonths !== null && (
                      <span className="text-[10px] text-slate-500 font-medium leading-tight mt-0.5">
                        {remainingMonths} mos left
                      </span>
                    )}
                  </div>
                </div>

                <span
                  className="px-2 py-0.5 rounded-full text-[10px] font-bold leading-none"
                  style={{
                    backgroundColor: style.badgeBackground,
                    color: style.badgeColor,
                  }}
                >
                  {progress}% Saved
                </span>
              </div>

              <div className="h-1.5 w-full rounded-full bg-[#EEE1E0] overflow-hidden">
                <div
                  className="h-full rounded-full transition-all"
                  style={{
                    width: `${progress}%`,
                    backgroundColor: style.progressColor,
                  }}
                />
              </div>

              <div className="flex items-center justify-between">
                <span className="text-[10px] font-semibold text-slate-600 bg-slate-100 px-1.5 py-0.5 rounded-md">
                  {formatMonthlyRequired(goal)}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </section>
  );
}

function getRemainingMonths(deadline: string): number | null {
  const end = new Date(deadline);

  if (Number.isNaN(end.getTime())) {
    return null;
  }

  const now = new Date();

  const months =
    (end.getFullYear() - now.getFullYear()) * 12 +
    (end.getMonth() - now.getMonth());

  return Math.max(0, months);
}

function formatMonthlyRequired(goal: DashboardGoal): string {
  // Replace with your actual monthly-contribution calculation/API field.
  if (goal.remaining <= 0) {
    return 'Goal reached';
  }

  const deadline = new Date(goal.deadline);
  const now = new Date();

  const months =
    (deadline.getFullYear() - now.getFullYear()) * 12 +
    (deadline.getMonth() - now.getMonth());

  if (months <= 0) {
    return 'Due now';
  }

  const monthly = Math.ceil(goal.remaining / months);

  return `₹${monthly.toLocaleString('en-IN')}/mo needed`;
}
