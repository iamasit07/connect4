import { Link } from "react-router-dom";
import { motion } from "framer-motion";
import { Gamepad2, Users, Bot, Trophy, ArrowRight, Github } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Header } from "@/components/layout/Header";
import { useAuthStore } from "@/features/auth/store/authStore";
import { BackgroundRippleEffect } from "@/components/ui/background-ripple-effect";

const LandingPage = () => {
  const { isAuthenticated } = useAuthStore();

  const fadeIn = {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    transition: { duration: 0.5 },
  };

  const features = [
    {
      icon: Users,
      title: "Global Competition",
      description: "Face opponents worldwide in real-time matches.",
    },
    {
      icon: Trophy,
      title: "Competitive Ranking",
      description: "Climb the Elo-based leaderboard with every victory.",
    },
    {
      icon: Bot,
      title: "Advanced AI",
      description: "Sharpen your skills against adaptive bot opponents.",
    },
  ];

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col font-sans selection:bg-primary/20">
      <Header />

      {/* Hero Section */}
      <section className="flex-1 flex flex-col justify-center items-center text-center px-4 py-32 relative overflow-hidden">
        {/* Background Ripple */}
        <div className="absolute inset-0 z-0">
          <BackgroundRippleEffect
            rows={20}
            cols={50}
            cellSize={40}
            className="[--cell-fill-color:hsl(0_0%_15%/_0.6)] [--cell-border-color:hsl(0_0%_20%/_0.3)]"
            style={{
              maskImage: `
                linear-gradient(to right, transparent, black 10%, black 70%, transparent),
                linear-gradient(to bottom, transparent, black 25%, black 75%, transparent)
              `,
              maskComposite: "intersect",
              WebkitMaskComposite: "source-in",
            }}
          />
        </div>

        <motion.div
          initial="initial"
          animate="animate"
          variants={fadeIn}
          className="max-w-6xl mx-auto space-y-8 relative z-10"
        >
          <div className="space-y-4">
            <h1 className="text-6xl sm:text-7xl md:text-8xl font-bold tracking-tight text-primary">
              Connect 4
            </h1>
            <p className="text-xl sm:text-2xl text-muted-foreground/80 font-light tracking-wide max-w-2xl mx-auto">
              Strategy. Simplicity. Mastery.
            </p>
          </div>

          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-8">
            {isAuthenticated ? (
              <Button
                size="lg"
                asChild
                className="h-14 px-8 text-lg rounded-full"
              >
                <Link to="/dashboard">
                  Enter Dashboard <ArrowRight className="ml-2 w-5 h-5" />
                </Link>
              </Button>
            ) : (
              <>
                <Button
                  size="lg"
                  asChild
                  className="h-14 px-10 text-lg rounded-full shadow-lg hover:shadow-primary/25 transition-all"
                >
                  <Link to="/login">Start Playing</Link>
                </Button>
              </>
            )}
          </div>
        </motion.div>
      </section>

      {/* Minimal Features Grid */}
      <section className="container py-24 border-t border-border/40">
        <div className="grid md:grid-cols-3 gap-12 px-4">
          {features.map((feature, i) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1 }}
              className="group space-y-4"
            >
              <div className="w-12 h-12 rounded-2xl bg-primary/5 flex items-center justify-center group-hover:bg-primary/10 transition-colors">
                <feature.icon
                  className="w-6 h-6 text-primary"
                  strokeWidth={1.5}
                />
              </div>
              <h3 className="text-xl font-semibold tracking-tight">
                {feature.title}
              </h3>
              <p className="text-muted-foreground leading-relaxed">
                {feature.description}
              </p>
            </motion.div>
          ))}
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border/40 py-12">
        <div className="container px-4 flex flex-col md:flex-row justify-between items-center gap-6 text-sm text-muted-foreground">
          <p>© 2025 Connect 4. Designed by Asit.</p>
          <div className="flex items-center gap-6">
            <a
              href="https://github.com/iamasit07"
              target="_blank"
              rel="noreferrer"
              className="hover:text-foreground transition-colors"
            >
              <Github className="w-5 h-5" />
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
};

export default LandingPage;
