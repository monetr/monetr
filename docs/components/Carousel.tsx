import * as React from 'react';
import { ArrowLeft, ArrowRight } from 'lucide-react';

import { Button } from '@monetr/docs/components/Button';
import mergeClasses from '@monetr/docs/util/mergeClasses';

import styles from './Carousel.module.scss';

import useEmblaCarousel, { type UseEmblaCarouselType } from 'embla-carousel-react';

type CarouselApi = UseEmblaCarouselType[1];
type UseCarouselParameters = Parameters<typeof useEmblaCarousel>;
type CarouselOptions = UseCarouselParameters[0];
type CarouselPlugin = UseCarouselParameters[1];

type CarouselProps = {
  opts?: CarouselOptions;
  plugins?: CarouselPlugin;
  orientation?: 'horizontal' | 'vertical';
  setApi?: (api: CarouselApi) => void;
};

type CarouselContextProps = {
  carouselRef: ReturnType<typeof useEmblaCarousel>[0];
  api: ReturnType<typeof useEmblaCarousel>[1];
  scrollPrev: () => void;
  scrollNext: () => void;
  canScrollPrev: boolean;
  canScrollNext: boolean;
} & CarouselProps;

const CarouselContext = React.createContext<CarouselContextProps | null>(null);

function useCarousel() {
  const context = React.useContext(CarouselContext);

  if (!context) {
    throw new Error('useCarousel must be used within a <Carousel />');
  }

  return context;
}

const Carousel = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement> & CarouselProps>(
  ({ orientation = 'horizontal', opts, setApi, plugins, className, children, ...props }, ref) => {
    const [carouselRef, api] = useEmblaCarousel(
      {
        ...opts,
        axis: orientation === 'horizontal' ? 'x' : 'y',
      },
      plugins,
    );
    const [canScrollPrev, setCanScrollPrev] = React.useState(false);
    const [canScrollNext, setCanScrollNext] = React.useState(false);

    const onSelect = React.useCallback((api: CarouselApi) => {
      if (!api) {
        return;
      }

      setCanScrollPrev(api.canScrollPrev());
      setCanScrollNext(api.canScrollNext());
    }, []);

    const scrollPrev = React.useCallback(() => {
      api?.scrollPrev();
    }, [api]);

    const scrollNext = React.useCallback(() => {
      api?.scrollNext();
    }, [api]);

    const handleKeyDown = React.useCallback(
      (event: React.KeyboardEvent<HTMLDivElement>) => {
        if (event.key === 'ArrowLeft') {
          event.preventDefault();
          scrollPrev();
        } else if (event.key === 'ArrowRight') {
          event.preventDefault();
          scrollNext();
        }
      },
      [scrollPrev, scrollNext],
    );

    React.useEffect(() => {
      if (!api || !setApi) {
        return;
      }

      setApi(api);
    }, [api, setApi]);

    React.useEffect(() => {
      if (!api) {
        return;
      }

      onSelect(api);
      api.on('reInit', onSelect);
      api.on('select', onSelect);

      return () => {
        api?.off('select', onSelect);
      };
    }, [api, onSelect]);

    return (
      <CarouselContext.Provider
        value={{
          carouselRef,
          api: api,
          opts,
          orientation: orientation || (opts?.axis === 'y' ? 'vertical' : 'horizontal'),
          scrollPrev,
          scrollNext,
          canScrollPrev,
          canScrollNext,
        }}
      >
        {/** biome-ignore lint/a11y/useSemanticElements: This is how it was provided by shadcn */}
        <div
          aria-roledescription='carousel'
          className={mergeClasses(styles.root, className)}
          onKeyDownCapture={handleKeyDown}
          ref={ref}
          role='region'
          {...props}
        >
          {children}
        </div>
      </CarouselContext.Provider>
    );
  },
);
Carousel.displayName = 'Carousel';

const CarouselContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    const { carouselRef, orientation } = useCarousel();

    return (
      <div className={styles.contentWrap} ref={carouselRef}>
        <div
          className={mergeClasses(
            styles.content,
            orientation === 'horizontal' ? styles.contentHorizontal : styles.contentVertical,
            className,
          )}
          ref={ref}
          {...props}
        />
      </div>
    );
  },
);
CarouselContent.displayName = 'CarouselContent';

const CarouselItem = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    const { orientation } = useCarousel();

    return (
      // biome-ignore lint/a11y/useSemanticElements: This is how its provided by shadcn
      <div
        aria-roledescription='slide'
        className={mergeClasses(
          styles.item,
          orientation === 'horizontal' ? styles.itemHorizontal : styles.itemVertical,
          className,
        )}
        ref={ref}
        role='group'
        {...props}
      />
    );
  },
);
CarouselItem.displayName = 'CarouselItem';

const CarouselPrevious = React.forwardRef<HTMLButtonElement, React.ComponentProps<typeof Button>>(
  ({ className, variant = 'outline', size = 'icon', ...props }, ref) => {
    const { orientation, scrollPrev, canScrollPrev } = useCarousel();

    return (
      <Button
        className={mergeClasses(
          styles.navButton,
          orientation === 'horizontal' ? styles.navPrevHorizontal : styles.navPrevVertical,
          className,
        )}
        disabled={!canScrollPrev}
        onClick={scrollPrev}
        ref={ref}
        size={size}
        variant={variant}
        {...props}
      >
        <ArrowLeft className={styles.arrowIcon} />
        <span className={styles.srOnly}>Previous slide</span>
      </Button>
    );
  },
);
CarouselPrevious.displayName = 'CarouselPrevious';

const CarouselNext = React.forwardRef<HTMLButtonElement, React.ComponentProps<typeof Button>>(
  ({ className, variant = 'outline', size = 'icon', ...props }, ref) => {
    const { orientation, scrollNext, canScrollNext } = useCarousel();

    return (
      <Button
        className={mergeClasses(
          styles.navButton,
          orientation === 'horizontal' ? styles.navNextHorizontal : styles.navNextVertical,
          className,
        )}
        disabled={!canScrollNext}
        onClick={scrollNext}
        ref={ref}
        size={size}
        variant={variant}
        {...props}
      >
        <ArrowRight className={styles.arrowIcon} />
        <span className={styles.srOnly}>Next slide</span>
      </Button>
    );
  },
);
CarouselNext.displayName = 'CarouselNext';

export { Carousel, type CarouselApi, CarouselContent, CarouselItem, CarouselNext, CarouselPrevious };
