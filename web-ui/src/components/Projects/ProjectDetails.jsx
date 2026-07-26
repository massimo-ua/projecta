import React, { useState } from 'react';
import { Outlet, useNavigate, useLocation, useParams } from 'react-router-dom';
import HomeLayout from '../../Layout';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  PieChart,
  Boxes,
  FileText,
  DollarSign,
  Package,
  Menu,
  ChevronRight,
  Settings,
} from 'lucide-react';
import { cn } from '@/lib/utils';

export function ProjectDetails() {
  const navigate = useNavigate();
  const location = useLocation();
  const { projectId } = useParams();

  const navGroups = [
    {
      title: 'Taxonomy',
      items: [
        { key: 'categories', label: 'Categories', icon: PieChart },
        { key: 'types', label: 'Types', icon: Boxes },
        { key: 'settings', label: 'Settings', icon: Settings },
      ],
    },
    {
      title: 'Operations',
      items: [
        { key: 'total', label: 'Total', icon: FileText },
        { key: 'payments', label: 'Payments', icon: DollarSign },
        { key: 'assets', label: 'Assets', icon: Package },
      ],
    },
  ];

  const currentTab = location.pathname.split('/').pop() || 'payments';

  const handleSelect = (key) => {
    navigate(key);
  };

  return (
    <HomeLayout>
      <div className="flex flex-col md:flex-row gap-6 min-h-[85vh]">
        {/* Mobile Navigation Dropdown */}
        <div className="md:hidden w-full flex items-center justify-between pb-2 border-b">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="w-full justify-between">
                <span className="flex items-center gap-2 capitalize">
                  <Menu className="h-4 w-4" />
                  Menu: {currentTab}
                </span>
                <ChevronRight className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent className="w-56">
              {navGroups.map((group) => (
                <React.Fragment key={group.title}>
                  <div className="px-2 py-1.5 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                    {group.title}
                  </div>
                  {group.items.map((item) => {
                    const Icon = item.icon;
                    return (
                      <DropdownMenuItem
                        key={item.key}
                        onClick={() => handleSelect(item.key)}
                        className={cn("flex items-center gap-2 cursor-pointer", currentTab === item.key && "font-semibold bg-accent")}
                      >
                        <Icon className="h-4 w-4" />
                        <span>{item.label}</span>
                      </DropdownMenuItem>
                    );
                  })}
                </React.Fragment>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Desktop Sidebar Navigation */}
        <aside className="hidden md:flex flex-col w-56 shrink-0 border-r pr-4 space-y-6">
          {navGroups.map((group) => (
            <div key={group.title} className="space-y-2">
              <h4 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider px-3">
                {group.title}
              </h4>
              <nav className="space-y-1">
                {group.items.map((item) => {
                  const Icon = item.icon;
                  const isActive = currentTab === item.key;
                  return (
                    <button
                      key={item.key}
                      onClick={() => handleSelect(item.key)}
                      className={cn(
                        "w-full flex items-center gap-3 px-3 py-2 text-sm font-medium rounded-lg transition-colors text-left",
                        isActive
                          ? "bg-primary text-primary-foreground shadow-sm"
                          : "text-muted-foreground hover:bg-accent hover:text-foreground"
                      )}
                    >
                      <Icon className="h-4 w-4 shrink-0" />
                      <span>{item.label}</span>
                    </button>
                  );
                })}
              </nav>
            </div>
          ))}
        </aside>

        {/* Main Content Area */}
        <div className="flex-1 min-w-0">
          <Outlet />
        </div>
      </div>
    </HomeLayout>
  );
}
