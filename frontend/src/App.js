import './App.css';
import {useEffect, useState} from "react";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  Avatar,
  Badge,
  Box,
  Card,
  CardContent,
  CardHeader,
  Grid,
  LinearProgress,
  linearProgressClasses,
  Paper,
  styled,
  Tooltip,
  Typography
} from "@mui/material";
import moment from "moment";
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';


function App() {
  const [data, setData] = useState({Characters: {}, Bank: []})

  useEffect(() => {
    const es = new EventSource("/events")

    es.onopen = () => console.log("SSE connection opened")

    es.onerror = (e) => console.log("SSE error", e)

    es.onmessage = (e) => {
      if (e.data) {
        const parsed = JSON.parse(e.data);

        setData(prevData => {
          const newData = {...prevData}
          if (parsed.Character) {
            newData.Characters = {...newData.Characters}
            newData.Characters[parsed.Character.Name] = parsed.Character
          }
          if (parsed.Bank) {
            newData.Bank = parsed.Bank
          }
          return newData
        })

      }
    }

    return () => es.close();
  }, []);

  const sortedCharNames = Object.keys(data.Characters).sort((a, b) => a.localeCompare(b))

  const numChars = sortedCharNames.length;
  return (
    <div className="App">
      <Box sx={{flexGrow: 1}} m={{xs: 0, sm: 2}}>
        <Grid container spacing={2}
              columns={{xl: numChars, lg: Math.min(numChars, 3), sm: Math.min(numChars, 2), xs: 1}}
              justifyContent="center" alignItems="stretch" direction="row">
          {sortedCharNames.map(charName =>
            <Grid item xs={1} sx={{display: "flex", flexDirection: "column"}} key={charName}>
              <Character char={data.Characters[charName]}/>
            </Grid>
          )}
          <Grid xs={12} item container spacing={2}>
            <Grid item xs={12} md={6} sx={{display: "flex", flexDirection: "column"}}>
              <Paper sx={{p: 2, height: "100%"}} elevation={4}>
                <Typography sx={{mb: 2}}>Bank</Typography>
                <Grid container spacing={2} justifyContent="space-evenly">
                  <ItemMap items={data.Bank}/>
                </Grid>
              </Paper>
            </Grid>
            <Grid item xs={12} md={6} sx={{display: "flex", flexDirection: "column"}}>
              <Paper sx={{p: 2, height: "100%"}} elevation={4}>
                <Typography sx={{mb: 2}}>Controls</Typography>
                <Typography variant="body">Coming soon!</Typography>
              </Paper>
            </Grid>
          </Grid>
        </Grid>
      </Box>
    </div>
  );
}

const skills = ["combat", "mining", "woodcutting", "fishing", "weaponcrafting", "gearcrafting", "jewelrycrafting", "cooking"]

const BorderLinearProgress = styled(LinearProgress)(({theme}) => ({
  height: 10,
  borderRadius: 5,
  [`&.${linearProgressClasses.colorPrimary}`]: {
    backgroundColor: theme.palette.grey[theme.palette.mode === 'light' ? 200 : 700],
  },
  [`& .${linearProgressClasses.bar}`]: {
    borderRadius: 5,
    backgroundColor: theme.palette.mode === 'light' ? '#1a90ff' : '#308fe8',
  },
}));

function LinearProgressWithLabel(props) {
  return (
    <Box sx={{display: 'flex', alignItems: 'center', flexGrow: 1}}>
      <Box sx={{width: '100%', mr: 1, flexGrow: 1}}>
        <BorderLinearProgress variant="determinate" {...props} />
      </Box>
    </Box>
  );
}

function buildStateString(state) {
  if (!state) {
    return ""
  }
  return state.map((state, i) => {
    return "  ".repeat(i) + state
  }).join("\n")
}

function Character({char}) {
  const [cooldownProgress, setCooldownProgress] = useState(0)

  useEffect(() => {
    const cooldownExpires = moment(char.CooldownExpires)

    const timer = setInterval(() => {
      const now = moment();
      const diffSeconds = cooldownExpires.diff(now) / 1000;
      const newProgress = diffSeconds / char.CooldownDuration * 100

      setCooldownProgress(newProgress)
    }, 500)

    return () => {
      clearInterval(timer)
    }
  }, [char.CooldownExpires]);

  const inventoryUsed = Object.values(char.Inventory).reduce((a, b) => a + b, 0)

  return <Card raised={true} elevation={4} sx={{height: "100%"}}>
    <LinearProgress variant="determinate" value={cooldownProgress}/>
    <CardHeader
      avatar={<Avatar src={`https://artifactsmmo.com/images/characters/${char.Skin}.png`}/>}
      title={char.Name}
      subheader={buildStateString(char.State)}
      subheaderTypographyProps={{whiteSpace: "pre-wrap"}}
    />
    <CardContent sx={{p: {xs: 1, sm: 2, xl: 2}}}>
      {char.Task &&
        <Box py={0} mb={2}>
          <Typography textTransform="capitalize">
            Task: {char.Task} ({char.TaskProgress} / {char.TaskTotal})
          </Typography>
          <Box mt={1} sx={{display: 'flex'}}>
            <Avatar sx={{mr: 2}} src={`https://artifactsmmo.com/images/monsters/${char.Task}.png`}/>
            <LinearProgressWithLabel
              variant="determinate"
              value={char.TaskProgress / char.TaskTotal * 100}/>
          </Box>
        </Box>
      }
      <Accordion>
        <AccordionSummary expandIcon={<ExpandMoreIcon/>}>
          <Typography sx={{width: '40%', flexShrink: 0}}>
            Skills
          </Typography>
          <Typography sx={{color: 'text.secondary'}}>
            {char.Gold} gold
          </Typography>
        </AccordionSummary>
        <AccordionDetails>
          <Grid container>
            {skills.map(skill =>
              <Grid item xs={6} lg={12} xl={6} sm={12} key={skill}>
                <Box py={0} mb={1}>
                  <Typography
                    textTransform="capitalize">{char.Levels[skill]} {skill} ({Math.round(char.Xp[skill] / char.MaxXp[skill] * 100)}%)</Typography>
                  <LinearProgressWithLabel variant="determinate"
                                           value={Math.round(char.Xp[skill] / char.MaxXp[skill] * 100)}/>
                </Box>
              </Grid>
            )}
          </Grid>
        </AccordionDetails>
      </Accordion>
      <Accordion defaultExpanded>
        <AccordionSummary expandIcon={<ExpandMoreIcon/>}>
          <Typography sx={{width: '40%', flexShrink: 0}}>
            Inventory
          </Typography>
          <Typography sx={{color: 'text.secondary'}}>
            {inventoryUsed} / {char.InventoryMaxItems} items
          </Typography>
        </AccordionSummary>
        <AccordionDetails>
          <Grid container spacing={1.5} justifyContent="center">
            <ItemMap items={char.Inventory}/>
          </Grid>
        </AccordionDetails>
      </Accordion>
      <Accordion defaultExpanded>
        <AccordionSummary expandIcon={<ExpandMoreIcon/>}>
          Equipment
        </AccordionSummary>
        <AccordionDetails>
          <Grid container spacing={1.5} justifyContent="center">
            {Object.keys(char.Equipment).map(slot =>
              <Grid item xs="auto" key={slot}>
                <Tooltip title={char.Equipment[slot]}>
                  <a href={`https://artifactsmmo.com/encyclopedia/items/${char.Equipment[slot]}`} target="_blank">
                    <Paper sx={{p: 1, pt: 1.5}} elevation={4}>
                      <Avatar src={getEquipmentIconUrl(char.Equipment, slot)} variant="rounded">
                        {slot}
                      </Avatar>
                    </Paper>
                  </a>
                </Tooltip>
              </Grid>
            )}
          </Grid>
        </AccordionDetails>
      </Accordion>
    </CardContent>
  </Card>
}

function ItemMap({items}) {
  return <>
    {Object.keys(items).map(itemCode =>
      <Grid item xs="auto" key={itemCode}>
        <Tooltip title={itemCode}>
          <a href={`https://artifactsmmo.com/encyclopedia/items/${itemCode}`} target="_blank">
            <Paper sx={{p: 1, pr: 2, pt: 1.5}} elevation={4}>
              <Badge badgeContent={items[itemCode]} max={999999} color="primary" overlap="circular">
                <Avatar src={`https://artifactsmmo.com/images/items/${itemCode}.png`} variant="rounded"/>
              </Badge>
            </Paper>
          </a>
        </Tooltip>
      </Grid>
    )}
  </>
}

function getEquipmentIconUrl(equipment, slot) {
  const item = equipment[slot]
  if (!!item) {
    return `https://artifactsmmo.com/images/items/${item}.png`
  }

  let osrsSlot = ""
  switch (slot) {
    case "weapon":
      osrsSlot = "Weapon_slot"
      break
    case "helmet":
      osrsSlot = "Head_slot"
      break
    case "amulet":
      osrsSlot = "Neck_slot"
      break
    case "body_armor":
      osrsSlot = "Body_slot"
      break
    case "shield":
      osrsSlot = "Shield_slot"
      break
    case "ring1":
      osrsSlot = "Ring_slot"
      break
    case "ring2":
      osrsSlot = "Ring_slot"
      break
    case "leg_armor":
      osrsSlot = "Legs_slot"
      break
    case "boots":
      osrsSlot = "Feet_slot"
      break
  }

  return `https://oldschool.runescape.wiki/images/${osrsSlot}.png`
}

export default App;
