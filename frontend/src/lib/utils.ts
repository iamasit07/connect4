import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const isInAppBrowser = () => {
  const userAgent = navigator.userAgent || navigator.vendor || (window as any).opera;
  // Common in-app browser identifiers
  const rules = [
    'LinkedInApp', // LinkedIn
    'FBAV', // Facebook
    'Instagram', // Instagram
    'Twitter', // Twitter
    'Line', // Line
    'Snapchat', // Snapchat
    'MicroMessenger', // WeChat
    'Frios', // Facebook iOS
    'Threads',
  ];
  return rules.some((rule) => userAgent.includes(rule));
};
