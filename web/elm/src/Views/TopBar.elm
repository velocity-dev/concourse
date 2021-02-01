module Views.TopBar exposing
    ( breadcrumbs
    , concourseLogo
    )

import Application.Models exposing (Session)
import Assets
import ColorValues
import Concourse exposing (hyphenNotation)
import Dashboard.FilterBuilder exposing (instanceGroupFilter)
import Html exposing (Html)
import Html.Attributes
    exposing
        ( class
        , href
        , id
        )
import Message.Message exposing (DomID(..), Message(..))
import RemoteData
import Routes
import SideBar.SideBar exposing (byPipelineId, lookupPipeline)
import Url
import Views.InstanceGroupBadge as InstanceGroupBadge
import Views.Styles as Styles


concourseLogo : Html Message
concourseLogo =
    Html.a (href "/" :: Styles.concourseLogo) []


breadcrumbs : Session -> Routes.Route -> Html Message
breadcrumbs session route =
    Html.div
        (id "breadcrumbs" :: Styles.breadcrumbContainer)
    <|
        case route of
            Routes.Pipeline { id } ->
                case lookupPipeline (byPipelineId id) session of
                    Nothing ->
                        []

                    Just pipeline ->
                        pipelineBreadcrumbs session pipeline

            Routes.Build { id } ->
                case lookupPipeline (byPipelineId id) session of
                    Nothing ->
                        []

                    Just pipeline ->
                        pipelineBreadcrumbs session pipeline
                            ++ [ breadcrumbSeparator
                               , jobBreadcrumb id.jobName
                               ]

            Routes.Resource { id } ->
                case lookupPipeline (byPipelineId id) session of
                    Nothing ->
                        []

                    Just pipeline ->
                        pipelineBreadcrumbs session pipeline
                            ++ [ breadcrumbSeparator
                               , resourceBreadcrumb id.resourceName
                               ]

            Routes.Job { id } ->
                case lookupPipeline (byPipelineId id) session of
                    Nothing ->
                        []

                    Just pipeline ->
                        pipelineBreadcrumbs session pipeline
                            ++ [ breadcrumbSeparator
                               , jobBreadcrumb id.jobName
                               ]

            Routes.Dashboard _ ->
                [ clusterNameBreadcrumb session ]

            _ ->
                []


breadcrumbComponent :
    { icon :
        { component : Assets.ComponentType
        , widthPx : Float
        , heightPx : Float
        }
    , name : String
    }
    -> List (Html Message)
breadcrumbComponent { name, icon } =
    [ Html.div
        (Styles.breadcrumbComponent icon)
        []
    , Html.text <| decodeName name
    ]


breadcrumbSeparator : Html Message
breadcrumbSeparator =
    Html.li
        (class "breadcrumb-separator" :: Styles.breadcrumbItem False)
        [ Html.text "/" ]


clusterNameBreadcrumb : Session -> Html Message
clusterNameBreadcrumb session =
    Html.div
        Styles.clusterName
        [ Html.text session.clusterName ]


pipelineBreadcrumbs : Session -> Concourse.Pipeline -> List (Html Message)
pipelineBreadcrumbs session pipeline =
    let
        pipelineGroup =
            session.pipelines
                |> RemoteData.withDefault []
                |> List.filter (\p -> p.name == pipeline.name && p.teamName == pipeline.teamName)

        inInstanceGroup =
            Concourse.isInstanceGroup pipelineGroup
    in
    (if inInstanceGroup then
        [ Html.a
            (id "breadcrumb-instance-group"
                :: (href <|
                        Routes.toString <|
                            Routes.Dashboard
                                { searchType = Routes.Normal <| instanceGroupFilter pipeline
                                , dashboardView = Routes.ViewNonArchivedPipelines
                                }
                   )
                :: Styles.breadcrumbItem True
            )
            [ InstanceGroupBadge.view ColorValues.white (List.length pipelineGroup)
            , Html.text pipeline.name
            ]
        , breadcrumbSeparator
        ]

     else
        []
    )
        ++ [ pipelineBreadcrumb inInstanceGroup pipeline ]


pipelineBreadcrumb : Bool -> Concourse.Pipeline -> Html Message
pipelineBreadcrumb inInstanceGroup pipeline =
    let
        text =
            if inInstanceGroup then
                hyphenNotation pipeline.instanceVars

            else
                pipeline.name
    in
    Html.a
        ([ id "breadcrumb-pipeline"
         , href <|
            Routes.toString <|
                Routes.pipelineRoute pipeline
         ]
            ++ Styles.breadcrumbItem True
        )
        (breadcrumbComponent
            { icon =
                { component = Assets.PipelineComponent
                , widthPx = 28
                , heightPx = 16
                }
            , name = text
            }
        )


jobBreadcrumb : String -> Html Message
jobBreadcrumb jobName =
    Html.li
        (id "breadcrumb-job" :: Styles.breadcrumbItem False)
        (breadcrumbComponent
            { icon =
                { component = Assets.JobComponent
                , widthPx = 32
                , heightPx = 17
                }
            , name = jobName
            }
        )


resourceBreadcrumb : String -> Html Message
resourceBreadcrumb resourceName =
    Html.li
        (id "breadcrumb-resource" :: Styles.breadcrumbItem False)
        (breadcrumbComponent
            { icon =
                { component = Assets.ResourceComponent
                , widthPx = 32
                , heightPx = 17
                }
            , name = resourceName
            }
        )


decodeName : String -> String
decodeName name =
    Maybe.withDefault name (Url.percentDecode name)
