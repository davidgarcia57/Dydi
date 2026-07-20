import React, { useEffect } from 'react';
import { Pressable } from 'react-native';
import Svg, { Circle, Ellipse, G, Line, Path, Rect } from 'react-native-svg';
import Animated, {
  useAnimatedStyle,
  useSharedValue,
  withRepeat,
  withSequence,
  withSpring,
  withTiming,
} from 'react-native-reanimated';
import WaterBottle from './WaterBottle';

// Ilustración grande por hábito — espejo del HabitHero del frontend web:
// escena "física" con formas rellenas y colores de marca. En móvil la escena
// respira con un loop suave de reanimated y hace un pop al tocarla (los loops
// por elemento del web se simplifican: una sola animación, cero jank).
const INK = '#2A251F';
const TERRACOTTA = '#C26F4D';

export default function HabitHero({
  iconKey = '',
  habitName = '',
  size = 150,
}: {
  iconKey?: string;
  habitName?: string;
  size?: number;
}) {
  const bob = useSharedValue(0);
  const pop = useSharedValue(1);

  useEffect(() => {
    bob.value = withRepeat(
      withSequence(withTiming(-5, { duration: 1600 }), withTiming(0, { duration: 1600 })),
      -1
    );
  }, [bob]);

  const animStyle = useAnimatedStyle(() => ({
    transform: [{ translateY: bob.value }, { scale: pop.value }],
  }));

  // Mismo hook robusto que el web: icon_key manda, el nombre rescata.
  const isWater = iconKey === 'water' || /agua|water/i.test(habitName || '');
  if (isWater) return <WaterBottle size={size} />;

  function onPress() {
    pop.value = withSequence(withTiming(0.93, { duration: 110 }), withSpring(1));
  }

  return (
    <Pressable onPress={onPress}>
      <Animated.View style={animStyle}>
        <Svg width={size} height={size * 1.2} viewBox="0 0 200 240">
          {/* Sombra de piso compartida */}
          <Ellipse cx={100} cy={212} rx={56} ry={8} fill={INK} opacity={0.08} />

          {/* Ejercicio: barra con discos */}
          {iconKey === 'exercise' && (
            <G>
              <Rect x={34} y={118} width={132} height={7} rx={3.5} fill={INK} />
              <Rect x={38} y={104} width={12} height={35} rx={5} fill="#A85B39" />
              <Rect x={150} y={104} width={12} height={35} rx={5} fill="#A85B39" />
              <Rect x={52} y={96} width={14} height={51} rx={5} fill="#C9714A" />
              <Rect x={134} y={96} width={14} height={51} rx={5} fill="#C9714A" />
              <Rect x={68} y={112} width={6} height={19} rx={2} fill={INK} />
              <Rect x={126} y={112} width={6} height={19} rx={2} fill={INK} />
            </G>
          )}

          {/* Pasos: tenis con líneas de velocidad */}
          {iconKey === 'steps' && (
            <G>
              <G stroke="#C9714A" strokeWidth={5} strokeLinecap="round" opacity={0.4}>
                <Line x1={18} y1={150} x2={40} y2={150} />
                <Line x1={10} y1={168} x2={36} y2={168} />
                <Line x1={20} y1={186} x2={42} y2={186} />
              </G>
              <Path
                d="M56 190 L56 156 Q56 146 66 143 Q92 136 108 122 Q116 115 124 119 Q146 132 160 136 Q170 139 170 148 L170 190 Z"
                fill="#C9714A"
              />
              <Path
                d="M56 168 Q80 158 104 166 Q136 176 170 170 L170 190 L56 190 Z"
                fill="#FCF9F3"
                opacity={0.35}
              />
              <Rect
                x={50}
                y={186}
                width={126}
                height={16}
                rx={8}
                fill="#FCF9F3"
                stroke={INK}
                strokeWidth={4}
              />
              <G stroke="#FCF9F3" strokeWidth={4} strokeLinecap="round">
                <Line x1={104} y1={136} x2={118} y2={146} />
                <Line x1={96} y1={146} x2={112} y2={156} />
                <Line x1={88} y1={156} x2={104} y2={166} />
              </G>
            </G>
          )}

          {/* Fruta: manzana con hoja */}
          {iconKey === 'fruit' && (
            <G>
              <Circle cx={100} cy={150} r={46} fill="#D96C57" />
              <Ellipse
                cx={82}
                cy={132}
                rx={14}
                ry={20}
                fill="#FCF9F3"
                opacity={0.28}
                transform="rotate(-24 82 132)"
              />
              <Path
                d="M100 106 q2 -16 12 -21"
                stroke={INK}
                strokeWidth={5}
                fill="none"
                strokeLinecap="round"
              />
              <Path d="M104 96 C110 78 128 72 142 76 C138 94 120 100 104 96 Z" fill="#A8C39A" />
            </G>
          )}

          {/* Sin chatarra: vaso con señal de prohibido */}
          {iconKey === 'no_sugar' && (
            <G>
              <Path d="M112 84 L126 42 L134 45 L120 84 Z" fill="#E9C281" />
              <Path
                d="M66 94 L134 94 L126 194 Q125 202 116 202 L84 202 Q75 202 74 194 Z"
                fill="#FCF9F3"
                stroke={INK}
                strokeWidth={4}
                strokeLinejoin="round"
              />
              <Rect x={60} y={82} width={80} height={13} rx={5} fill={TERRACOTTA} />
              <Rect x={72} y={124} width={56} height={10} rx={5} fill="#EDA48F" opacity={0.6} />
              <Rect x={74} y={152} width={52} height={10} rx={5} fill="#EDA48F" opacity={0.6} />
              <Circle cx={100} cy={140} r={66} fill="none" stroke="#BC5C42" strokeWidth={8} />
              <Line
                x1={54}
                y1={187}
                x2={146}
                y2={93}
                stroke="#BC5C42"
                strokeWidth={8}
                strokeLinecap="round"
              />
            </G>
          )}

          {/* Leer: libro abierto */}
          {iconKey === 'read' && (
            <G>
              <Path d="M28 152 Q100 130 172 152 L172 180 Q100 158 28 180 Z" fill={TERRACOTTA} />
              <Path
                d="M100 88 C80 74 52 72 32 78 L32 162 C52 156 80 158 100 172 Z"
                fill="#FCF9F3"
                stroke={INK}
                strokeWidth={4}
                strokeLinejoin="round"
              />
              <Path
                d="M100 88 C120 74 148 72 168 78 L168 162 C148 156 120 158 100 172 Z"
                fill="#FCF9F3"
                stroke={INK}
                strokeWidth={4}
                strokeLinejoin="round"
              />
              <Line x1={100} y1={88} x2={100} y2={172} stroke={INK} strokeWidth={4} />
              <G stroke="#A89C89" strokeWidth={3.5} strokeLinecap="round">
                <Line x1={44} y1={98} x2={84} y2={92} />
                <Line x1={44} y1={112} x2={84} y2={106} />
                <Line x1={44} y1={126} x2={76} y2={121} />
                <Line x1={116} y1={92} x2={156} y2={98} />
                <Line x1={116} y1={106} x2={156} y2={112} />
              </G>
              <Path d="M148 52 L150 57 L155 59 L150 61 L148 66 L146 61 L141 59 L146 57 Z" fill="#D4A847" />
            </G>
          )}

          {/* Foco: diana con flecha */}
          {iconKey === 'focus' && (
            <G>
              <G stroke={INK} strokeWidth={5} strokeLinecap="round">
                <Line x1={82} y1={182} x2={70} y2={208} />
                <Line x1={118} y1={182} x2={130} y2={208} />
              </G>
              <Circle cx={100} cy={126} r={56} fill="#FCF9F3" stroke={INK} strokeWidth={4} />
              <Circle cx={100} cy={126} r={40} fill="#EDA48F" />
              <Circle cx={100} cy={126} r={24} fill="#FCF9F3" />
              <Circle cx={100} cy={126} r={10} fill="#BC5C42" />
              <Line x1={54} y1={64} x2={97} y2={120} stroke={INK} strokeWidth={5} strokeLinecap="round" />
              <Path d="M54 64 L42 58 L50 52 Z" fill="#A8C39A" />
              <Path d="M60 72 L48 66 L56 60 Z" fill="#A8C39A" />
            </G>
          )}

          {/* Journaling: libreta con pluma */}
          {iconKey === 'journal' && (
            <G>
              <Rect x={52} y={60} width={96} height={136} rx={10} fill="#FCF9F3" stroke={INK} strokeWidth={4} />
              <G stroke={INK} strokeWidth={4} strokeLinecap="round">
                <Line x1={70} y1={54} x2={70} y2={68} />
                <Line x1={90} y1={54} x2={90} y2={68} />
                <Line x1={110} y1={54} x2={110} y2={68} />
                <Line x1={130} y1={54} x2={130} y2={68} />
              </G>
              <G stroke="#A89C89" strokeWidth={3.5} strokeLinecap="round">
                <Line x1={68} y1={92} x2={132} y2={92} />
                <Line x1={68} y1={112} x2={132} y2={112} />
                <Line x1={68} y1={132} x2={120} y2={132} />
              </G>
              <Line x1={68} y1={156} x2={126} y2={156} stroke={TERRACOTTA} strokeWidth={4} strokeLinecap="round" />
              <Rect
                x={120}
                y={130}
                width={10}
                height={30}
                rx={4}
                fill={TERRACOTTA}
                transform="rotate(38 125 145)"
              />
              <Path d="M112 158 L118 166 L108 164 Z" fill={INK} />
            </G>
          )}

          {/* Sin redes: burbuja de chat en pausa */}
          {iconKey === 'no_social' && (
            <G>
              <Rect x={60} y={120} width={80} height={50} rx={14} fill="#E4EDDC" />
              <Rect x={44} y={70} width={112} height={72} rx={16} fill="#FCF9F3" stroke={INK} strokeWidth={4} />
              <Path
                d="M66 140 L58 162 L86 142 Z"
                fill="#FCF9F3"
                stroke={INK}
                strokeWidth={4}
                strokeLinejoin="round"
              />
              <G fill="#7B8FA1">
                <Circle cx={76} cy={106} r={6} />
                <Circle cx={100} cy={106} r={6} />
                <Circle cx={124} cy={106} r={6} />
              </G>
              <Circle cx={100} cy={122} r={64} fill="none" stroke="#BC5C42" strokeWidth={8} />
              <Line
                x1={55}
                y1={167}
                x2={145}
                y2={77}
                stroke="#BC5C42"
                strokeWidth={8}
                strokeLinecap="round"
              />
            </G>
          )}

          {/* Sin teléfono de noche: teléfono bocabajo bajo luna y estrellas */}
          {iconKey === 'no_phone' && (
            <G>
              <Path d="M148 58 A26 26 0 1 1 118 34 A20 20 0 0 0 148 58 Z" fill="#F5E8CD" />
              <Path d="M56 52 L58 58 L64 60 L58 62 L56 68 L54 62 L48 60 L54 58 Z" fill="#D4A847" />
              <Path
                d="M84 34 L85.5 38.5 L90 40 L85.5 41.5 L84 46 L82.5 41.5 L78 40 L82.5 38.5 Z"
                fill="#D4A847"
              />
              <G transform="rotate(-6 100 166)">
                <Rect x={54} y={142} width={92} height={50} rx={10} fill={INK} />
                <Rect x={60} y={148} width={80} height={38} rx={6} fill="#3A342C" />
                <Circle cx={100} cy={188} r={2.5} fill="#FCF9F3" opacity={0.5} />
              </G>
              <G fill="none" stroke="#7B8FA1" strokeWidth={4} strokeLinecap="round" strokeLinejoin="round">
                <Path d="M124 122 h12 l-12 12 h12" />
                <Path d="M146 100 h9 l-9 9 h9" />
              </G>
            </G>
          )}

          {/* Tender la cama */}
          {iconKey === 'bed' && (
            <G>
              <Rect x={36} y={92} width={16} height={90} rx={7} fill={TERRACOTTA} />
              <Rect x={44} y={142} width={126} height={38} rx={11} fill="#FCF9F3" stroke={INK} strokeWidth={4} />
              <Path
                d="M84 142 L160 142 Q170 142 170 153 L170 168 Q170 180 158 180 L84 180 Q76 160 84 142 Z"
                fill="#A8C39A"
              />
              <Path d="M92 148 Q88 161 92 174" stroke="#5C7650" strokeWidth={3.5} fill="none" strokeLinecap="round" />
              <Rect x={52} y={126} width={36} height={22} rx={10} fill="#EFE7D8" stroke={INK} strokeWidth={3.5} />
              <G fill={INK}>
                <Rect x={48} y={180} width={8} height={14} rx={3} />
                <Rect x={158} y={180} width={8} height={14} rx={3} />
              </G>
              <Path d="M150 104 L153 112 L161 115 L153 118 L150 126 L147 118 L139 115 L147 112 Z" fill="#D4A847" />
              <Path d="M120 84 L122 89 L127 91 L122 93 L120 98 L118 93 L113 91 L118 89 Z" fill="#D4A847" />
            </G>
          )}

          {/* Aire libre: sol, lomas, árbol y nube */}
          {iconKey === 'outdoors' && (
            <G>
              <G stroke="#D4A847" strokeWidth={5} strokeLinecap="round">
                <Line x1={142} y1={34} x2={142} y2={44} />
                <Line x1={142} y1={100} x2={142} y2={110} />
                <Line x1={104} y1={72} x2={114} y2={72} />
                <Line x1={170} y1={72} x2={180} y2={72} />
                <Line x1={115} y1={45} x2={122} y2={52} />
                <Line x1={162} y1={92} x2={169} y2={99} />
                <Line x1={115} y1={99} x2={122} y2={92} />
                <Line x1={162} y1={52} x2={169} y2={45} />
              </G>
              <Circle cx={142} cy={72} r={20} fill="#D4A847" />
              <Path
                d="M42 80 Q42 68 54 68 Q58 56 72 58 Q84 58 86 70 Q96 72 94 82 Q92 90 82 90 L52 90 Q42 90 42 80 Z"
                fill="#FCF9F3"
                stroke={INK}
                strokeWidth={3.5}
              />
              <Path d="M4 210 Q52 148 116 174 Q166 192 198 180 L198 210 Z" fill="#A8C39A" opacity={0.55} />
              <Path d="M4 210 Q70 168 130 188 Q170 200 198 192 L198 210 Z" fill="#A8C39A" />
              <Rect x={58} y={152} width={9} height={32} rx={4} fill="#8A6B4F" />
              <G fill="#5C7650">
                <Circle cx={63} cy={138} r={17} />
                <Circle cx={49} cy={149} r={11} />
                <Circle cx={77} cy={149} r={11} />
              </G>
            </G>
          )}

          {/* Fallback: destello grande */}
          {![
            'exercise',
            'steps',
            'fruit',
            'no_sugar',
            'read',
            'focus',
            'journal',
            'no_social',
            'no_phone',
            'bed',
            'outdoors',
          ].includes(iconKey) && (
            <G>
              <Path
                d="M100 76 L108 118 L150 126 L108 134 L100 176 L92 134 L50 126 L92 118 Z"
                fill="#D4A847"
              />
              <Path d="M146 84 L149 92 L157 95 L149 98 L146 106 L143 98 L135 95 L143 92 Z" fill="#E9C281" />
            </G>
          )}
        </Svg>
      </Animated.View>
    </Pressable>
  );
}
