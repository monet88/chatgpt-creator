const FIRST_NAMES = [
  'james', 'john', 'robert', 'michael', 'william', 'david', 'richard', 'joseph',
  'thomas', 'charles', 'christopher', 'daniel', 'matthew', 'anthony', 'mark',
  'mary', 'patricia', 'jennifer', 'linda', 'barbara', 'elizabeth', 'susan',
  'jessica', 'sarah', 'karen', 'lisa', 'nancy', 'betty', 'margaret', 'sandra',
  'ashley', 'emily', 'olivia', 'emma', 'sophia', 'ava', 'isabella', 'mia',
  'liam', 'noah', 'ethan', 'mason', 'logan', 'lucas', 'jackson', 'aiden',
  'leon', 'kai', 'alex', 'sam', 'jordan', 'taylor', 'morgan', 'casey',
  'ryan', 'kevin', 'brian', 'george', 'edward', 'ronald', 'timothy', 'jason',
];

const LAST_NAMES = [
  'smith', 'johnson', 'williams', 'brown', 'jones', 'garcia', 'miller', 'davis',
  'rodriguez', 'martinez', 'hernandez', 'lopez', 'gonzalez', 'wilson', 'anderson',
  'thomas', 'taylor', 'moore', 'jackson', 'martin', 'lee', 'perez', 'thompson',
  'white', 'harris', 'sanchez', 'clark', 'ramirez', 'lewis', 'robinson',
  'walker', 'young', 'allen', 'king', 'wright', 'scott', 'torres', 'nguyen',
  'hill', 'flores', 'green', 'adams', 'nelson', 'baker', 'hall', 'rivera',
  'campbell', 'mitchell', 'carter', 'roberts', 'turner', 'phillips', 'evans',
  'parker', 'edwards', 'collins', 'stewart', 'morris', 'cook', 'rogers',
];

const pick = <T>(arr: T[]): T => arr[Math.floor(Math.random() * arr.length)];

const randomSuffix = () => {
  const bytes = crypto.getRandomValues(new Uint8Array(2));
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
};

export const randomLocalPart = (): string =>
  `${pick(FIRST_NAMES)}.${pick(LAST_NAMES)}.${randomSuffix()}`;
