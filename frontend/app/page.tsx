import { redirect } from "next/navigation";

/**
 * Home Page - Redirect to Insights
 * Automatically redirect root path to /insights
 */
export default function Home() {
  redirect("/insights");
}
