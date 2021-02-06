module Dashboard.Text exposing
    ( asciiArt
    , velocityIntroductions
    , setPipelineInstructions
    , welcome
    )


asciiArt : String
asciiArt =
    String.join "\n"
        [ "                          `:::                                             "
        , "                         `:::::                                            "
        , "                         :::::::                                           "
        , "                         ::::::::`                                         "
        , "                          ::::::::,           :                            "
        , "                           :::::::::      ::: ::                           "
        , "                            :::::::::    :::::` ,                          "
        , "                             :::::::::  :::::::                            "
        , "                              :::::::::::::::::`                           "
        , "                               ::::::::::::::::                            "
        , "                                ::::::::::::::.                            "
        , "                           `:`   ::::::, `:::.                             "
        , "                          `:.     ::::,  :::.                              "
        , "                      :: `:.      :::,  ::::                               "
        , "                     :: `:.      ::::  ::::::                              "
        , "                    :: `:.      ,:::::::::::::                             "
        , "                   ::  :.      .:::::::::::::::                            "
        , "                  ,:           ::::::::::::::::.                           "
        , "                              ::::::::. ::::::::`                          "
        , "                             ::::::::`   ::::::::                `         "
        , "                            ::::::::      ::::::::               ::`       "
        , "                           ,:::::::        ::::::::              ,::,  . ` "
        , "                         :::::::::          ::::::::              ,:::,::  "
        , "                        ::::::::.            :::::::.              ,:::::  "
        , "                       ::::::::`              :::::::             ` :: :   "
        , "                      .:::::::                 :::::          `  : .:,::,  "
        , "                       .::::::            :.    :::          `     :::,::. "
        , "                        .:::::      .:   :.      .            .   :::  ,::`"
        , "                         .:::      .:   :.                   ,  ,::,    ,::"
        , "                          .,      .:   :.                   ,   :::      ,:"
        , "                                 .:   :.                         :  .      "
        , "                                `:   :.                            ` :     "
        , "                                                                    :      "
        , "    .                                                              `       "
        , "    ::                                                                     "
        , "     ::,:                                                                  "
        , "      : :                                                                  "
        , "     `:::                                                                  "
        , "     :, ::                                                                 "
        , "   .:.   :,                                                                "
        , "    :                                                                      "
        , "        `                                                                  "
        , "       .                                                                   "
        ]


welcome : String
welcome =
    "Welcome to Velocity!"


velocityIntroductions : String
velocityIntroductions =
    "Velocity automates and accelerates modern software delivery allowing your development teams to focus on features. Bring the power of agility and visualizations to your delivery pipelines."


setPipelineInstructions : String
setPipelineInstructions =
    "then, use `fly set-pipeline` to set up your new pipeline"
